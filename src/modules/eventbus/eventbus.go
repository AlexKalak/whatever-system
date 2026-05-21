package eventbus

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Topic[T any] struct {
	name string
}

func NewTopic[T any](name string) Topic[T] {
	return Topic[T]{name: name}
}

func (t Topic[T]) Name() string {
	return t.name
}

type Event[T any] struct {
	Topic     string    `json:"topic"`
	Timestamp time.Time `json:"timestamp"`
	Payload   T         `json:"payload"`
}

type Sender interface {
	Send(ctx context.Context, event Event[any]) error
}

type Handler func(ctx context.Context, event Event[any])

type EventBus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

func NewEventBus() *EventBus {
	return &EventBus{handlers: map[string][]Handler{}}
}

func (b *EventBus) Subscribe(topic string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[topic] = append(b.handlers[topic], handler)
}

func SubscribeTopic[T any](b *EventBus, topic Topic[T], handler func(ctx context.Context, event Event[T])) {
	b.Subscribe(topic.Name(), func(ctx context.Context, event Event[any]) {
		payload, ok := event.Payload.(T)
		if !ok {
			return
		}

		handler(ctx, Event[T]{
			Topic:     event.Topic,
			Timestamp: event.Timestamp,
			Payload:   payload,
		})
	})
}

func (b *EventBus) Send(ctx context.Context, event Event[any]) error {
	b.mu.RLock()
	handlers := append([]Handler(nil), b.handlers[event.Topic]...)
	b.mu.RUnlock()

	for _, handler := range handlers {
		handler(ctx, event)
	}

	return nil
}

func Publish[T any](ctx context.Context, s Sender, topic Topic[T], payload T, ts time.Time) error {
	if ts.IsZero() {
		ts = time.Now()
	}

	return s.Send(ctx, Event[any]{
		Topic:     topic.Name(),
		Timestamp: ts,
		Payload:   payload,
	})
}

func MustPayload[T any](event Event[any]) (T, error) {
	payload, ok := event.Payload.(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("unexpected payload type for topic=%s", event.Topic)
	}
	return payload, nil
}
