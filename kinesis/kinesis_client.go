package kinesis

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
)

type Client struct {
	Kinesis *kinesis.Kinesis
}

type clientController struct {
	k   *kinesis.Kinesis
	err error
}

// NewClient creates a client to manage Kinesis connection.
// It sets new configuration on AWS and passes Credentials to the service.
// If no credentials are provided the default method WithEnvironmentVariables
// will be used.
func NewClient(ctx context.Context, creds ...Credentials) (*Client, error) {
	newConfig := aws.NewConfig()
	switch len(creds) {
	case 0:
		newConfig.WithCredentials(credentials.NewCredentials(WithEnvironmentVariables()))
	case 1:
		newConfig.WithCredentials(credentials.NewCredentials(creds[0]))
	default:
		return nil, fmt.Errorf("creds length should be 0 or 1 not %v", len(creds))
	}
	controller := make(chan clientController, 1)
	go func() {
		s, err := session.NewSession(newConfig)
		if err != nil {
			controller <- clientController{
				err: fmt.Errorf("new aws session error: %v", err),
			}
			return
		}
		controller <- clientController{
			k: kinesis.New(s),
		}
	}()
	select {
	case <-ctx.Done():
		return nil, nil
	case ctrl := <-controller:
		if ctrl.err != nil {
			return nil, ctrl.err
		}
		return &Client{Kinesis: ctrl.k}, nil
	}
}
