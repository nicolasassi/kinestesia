package pubsub

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"github.com/nicolasassi/kinestesia/translator"
	"google.golang.org/api/option"
)

type Client struct {
	client *pubsub.Client
	name string
	topics []string
	// Translator represents how should the incoming data be in the end of the process.
	// If Translator is nil the data will go as it came to the receiver.
	translator *translator.Translator
	stream chan []byte
	sent chan struct{}
	errors chan error
}

func NewPubSubClient(ctx context.Context, projectID string, opts ...option.ClientOption) (*Client, error) {
	client, err := pubsub.NewClient(ctx, projectID, opts...)
	if err != nil {
		return nil, fmt.Errorf("new pubsub client error: %v", err)
	}
	return &Client{
		client: client,
		name: "pubsub",
		stream: make(chan []byte),
		sent: make(chan struct{}),
		errors: make(chan error, 1),
	}, nil
}

func (c Client) String() string {
	return c.name
}

func (c *Client) AddTopics(topics ...string) {
	c.topics = append(c.topics, topics...)
}

func (c *Client) AddMessage(b []byte) {
	c.stream <- b
	<-c.sent
}

func (c Client) TranslationRequired() bool {
	return c.translator != nil
}

// SetTranslation is a setter for translation.
// The keys are paths to get to the data in the coming JSON and teh value should be
// the key which are going to be associated with this data in the response JSON.
// The key paths should be separated by dot if the data is in a inner layer of the JSON.
// ex: map["payload.contact.name"] = "name"
// If indexing is required while translating a path this should be done as follows:
// ex: map["payload.contacts.[0].name"] = "first_contact_name".
// Only translated fields will be included in the final response, so even if no actual translation
// is required the field name should be added:
// ex: map["payload"] = "payload"
func (c *Client) SetTranslation(t *translator.Translator) {
	c.translator = t
}

func (c *Client) Translate(b []byte) ([]byte, error) {
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	resp := c.translator.Translate(m)
	if resp == nil {
		return nil, nil
	}
	bb, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	return bb, nil
}

func (c *Client) Send(ctx context.Context) error {
	var topics []*pubsub.Topic
	for _, topicID := range c.topics {
		topic := c.client.Topic(topicID)
		topics = append(topics, topic)
	}

	results := make(chan *pubsub.PublishResult)
	go c.watch(ctx, results)
	for {
		select {
		case <-ctx.Done():
			for _, topic := range topics {
				topic.Stop()
			}
			return nil
		case err := <-c.errors:
			return err
		case message := <-c.stream:
			for _, topic := range topics {
				r := topic.Publish(ctx, &pubsub.Message{
					Data: message,
				})
				results <- r
				c.sent <- struct{}{}
			}
		}
	}
}

func (c *Client) watch(ctx context.Context, results chan *pubsub.PublishResult) {
	for result := range results {
		go func(result *pubsub.PublishResult) {
			_, err := result.Get(ctx)
			if err == context.Canceled {
				return
			}
			if err != nil {
				c.errors <- fmt.Errorf("[PUBLISH]: %v", err)
				return
			}
		}(result)
	}
}
