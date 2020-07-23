package kinesis

import (
	"context"
	"fmt"
	consumer "github.com/harlow/kinesis-consumer"
	"github.com/nicolasassi/kinestesia/receivers"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"sync"
)

const (
	maxWorkersForReceivers = 20
)

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

func NewStreamer(ctx context.Context, streamName string, opts ...*Client) (*Streamer, error) {
	var client *Client
	switch len(opts) {
	case 0:
		c, err := NewClient(ctx)
		if err != nil {
			return nil, err
		}
		client = c
	case 1:
		client = opts[0]
	default:
		return nil, fmt.Errorf("opts should be 1 or 0 not %v", len(opts))
	}
	controller := make(chan streamController, 1)
	go func() {
		c, err := consumer.New(streamName, consumer.WithClient(client.Kinesis))
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
		go func(rec receivers.Receiver) {
			errChan <- rec.Send(ctx)
		}(rec)
	}
	sem := semaphore.NewWeighted(int64(maxWorkersForReceivers))
	go func() {
		wg := new(sync.WaitGroup)
		errChan <- s.c.Scan(ctx, func(r *consumer.Record) error {
			for _, rec := range args {
				if err := sem.Acquire(ctx, 1); err != nil {
					errChan <- err
					break
				}
				wg.Add(1)
				go func(rec receivers.Receiver, data []byte) {
					defer wg.Done()
					defer sem.Release(1)
					if rec.TranslationRequired() {
						translated, err := rec.Translate(data)
						if err != nil {
							errChan <- fmt.Errorf("receiver service %s error: %v", rec.String(), err)
						}
						rec.AddMessage(translated)
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
		return nil
	case err := <-errChan:
		return err
	}
}

type Streamers []*Streamer

func NewStreamers(ctx context.Context, args ...interface{}) (*Streamers, error) {
	var streamNames []string
	var client *Client
	var streamers Streamers
	for _, arg := range args {
		switch arg.(type) {
		case string:
			streamNames = append(streamNames, arg.(string))
		case *Client:
			client = arg.(*Client)
		}
	}
	g := new(errgroup.Group)
	locker := new(sync.Mutex)
	for _, streamName := range streamNames {
		func(streamName string) {
			g.Go(func() error {
				if client != nil {
					streamer, err := NewStreamer(ctx, streamName, client)
					if err != nil {
						return err
					}
					locker.Lock()
					defer locker.Unlock()
					streamers = append(streamers, streamer)
				}
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
