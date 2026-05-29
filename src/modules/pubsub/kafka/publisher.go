package kafka

import (
	"context"

	"github.com/alexkalak/whatever-system/src/modules/pubsub"
	segmentio "github.com/segmentio/kafka-go"
)

type PublisherConfig struct {
	Brokers []string
}

type Publisher struct {
	writer *segmentio.Writer
}

func NewPublisher(cfg PublisherConfig) *Publisher {
	return &Publisher{
		writer: &segmentio.Writer{
			Addr:     segmentio.TCP(cfg.Brokers...),
			Balancer: &segmentio.Hash{},
		},
	}
}

func (p *Publisher) Publish(ctx context.Context, msg pubsub.Message) error {
	headers := make([]segmentio.Header, 0, len(msg.Headers))
	for key, value := range msg.Headers {
		headers = append(headers, segmentio.Header{Key: key, Value: []byte(value)})
	}

	return p.writer.WriteMessages(ctx, segmentio.Message{
		Topic:   msg.Topic,
		Key:     msg.Key,
		Value:   msg.Value,
		Headers: headers,
	})
}

func (p *Publisher) Close() error {
	return p.writer.Close()
}

var _ pubsub.Publisher = (*Publisher)(nil)
