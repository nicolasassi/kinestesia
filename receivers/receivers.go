package receivers

import (
	"context"
)

type Receiver interface {
	Send(ctx context.Context) error
	AddMessage(b []byte)
	Translate(b []byte) ([]byte, error)
	TranslationRequired() bool
	String() string
}


