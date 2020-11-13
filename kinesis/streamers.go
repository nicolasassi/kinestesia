package kinesis

import (
	"context"
	"fmt"
	consumer "github.com/harlow/kinesis-consumer"
	"github.com/nicolasassi/kinestesia/receivers"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"log"
	"sync"
)

const (
	maxWorkersForReceivers = 20
)

var once sync.Once

type Streaming interface {
	Stream(ctx context.Context, receivers ...receivers.Receiver) error
}

type Streamer struct {
	c *consumer.Consumer
}

type streamController struct {
	c   *consumer.Consumer
	err error
}

func NewStreamer(ctx context.Context, streamName string, opts ...consumer.Option) (*Streamer, error) {
	controller := make(chan streamController, 1)
	go func() {
		c, err := consumer.New(streamName, opts...)
		if err != nil {
			controller <- streamController{
				err: fmt.Errorf("new consumer error: %v", err),
			}
			return
		}
		controller <- streamController{
			c: c,
		}
	}()
	select {
	case <-ctx.Done():
		return nil, nil
	case ctrl := <-controller:
		if ctrl.err != nil {
			return nil, ctrl.err
		}
		return &Streamer{c: ctrl.c}, nil
	}
}

func (s *Streamer) Stream(ctx context.Context, args ...receivers.Receiver) error {
	errChan := make(chan error, 1)
	for _, rec := range args {
		once.Do(func() {
			go func(rec receivers.Receiver) {
				errChan <- rec.Send(ctx)
			}(rec)
		})
	}
	go func() {
		sem := semaphore.NewWeighted(int64(maxWorkersForReceivers))
		errChan <- s.c.Scan(ctx, func(r *consumer.Record) error {
			log.Println("Go here once more")
			for _, rec := range args {
				if err := sem.Acquire(ctx, 1); err != nil {
					errChan <- err
					break
				}
				go func(rec receivers.Receiver, data []byte) {
					log.Println("And here once more")
					defer sem.Release(1)
					if rec.TranslationRequired() {
						translated, err := rec.Translate(data)
						if err != nil {
							errChan <- fmt.Errorf("receiver service %s error: %v", rec.String(), err)
							return
						}
						if translated != nil {
							rec.AddMessage(translated)
						}
						return
					}
					rec.AddMessage(data)
				}(rec, r.Data)
			}
			return nil // continue scanning
		})
	}()
	select {
	case <-ctx.Done():
		log.Printf("Done Stream")
		return nil
	case err := <-errChan:
		return err
	}
}

type Streamers []*Streamer

func NewStreamers(ctx context.Context, args ...interface{}) (*Streamers, error) {
	var streamNames []string
	var opts []consumer.Option
	var streamers Streamers
	for _, arg := range args {
		switch arg.(type) {
		case string:
			if arg.(string) == "" {
				continue
			}
			streamNames = append(streamNames, arg.(string))
		case consumer.Option:
			opts = append(opts, arg.(consumer.Option))
		}
	}
	g := new(errgroup.Group)
	locker := new(sync.Mutex)
	for _, streamName := range streamNames {
		func(streamName string) {
			g.Go(func() error {
				streamer, err := NewStreamer(ctx, streamName, opts...)
				if err != nil {
					return err
				}
				locker.Lock()
				defer locker.Unlock()
				streamers = append(streamers, streamer)
				return nil
			})
		}(streamName)
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return &streamers, nil
}

func (ss *Streamers) Stream(ctx context.Context, receivers ...receivers.Receiver) error {
	g := new(errgroup.Group)
	for _, streamer := range *ss {
		func(streamer *Streamer) {
			g.Go(func() error {
				if err := streamer.Stream(ctx, receivers...); err != nil {
					return err
				}
				return nil
			})
		}(streamer)
	}
	if err := g.Wait(); err != nil {
		return err
	}
	return nil
}
