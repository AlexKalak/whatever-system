package pubsub

import "context"

type Message struct {
	Topic   string
	Key     []byte
	Value   []byte
	Headers map[string]string
	AckFunc func(context.Context) error
}

func (m Message) Ack(ctx context.Context) error {
	if m.AckFunc == nil {
		return nil
	}
	return m.AckFunc(ctx)
}

type Publisher interface {
	Publish(ctx context.Context, msg Message) error
	Close() error
}

type Handler func(ctx context.Context, msg Message) error

type Consumer interface {
	Subscribe(ctx context.Context, handler Handler) error
	Close() error
}
