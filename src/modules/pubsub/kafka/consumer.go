package kafka

import (
	"context"

	"github.com/alexkalak/whatever-system/src/modules/pubsub"
	segmentio "github.com/segmentio/kafka-go"
)

type ConsumerConfig struct {
	Brokers []string
	Topic   string
	GroupID string
}

type Consumer struct {
	reader *segmentio.Reader
}

func NewConsumer(cfg ConsumerConfig) *Consumer {
	return &Consumer{
		reader: segmentio.NewReader(segmentio.ReaderConfig{
			Brokers: cfg.Brokers,
			Topic:   cfg.Topic,
			GroupID: cfg.GroupID,
		}),
	}
}

func (c *Consumer) Subscribe(ctx context.Context, handler pubsub.Handler) error {
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			return err
		}

		headers := make(map[string]string, len(msg.Headers))
		for _, header := range msg.Headers {
			headers[header.Key] = string(header.Value)
		}

		if err := handler(ctx, pubsub.Message{
			Topic:   msg.Topic,
			Key:     msg.Key,
			Value:   msg.Value,
			Headers: headers,
			AckFunc: func(ctx context.Context) error {
				return c.reader.CommitMessages(ctx, msg)
			},
		}); err != nil {
			return err
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}

var _ pubsub.Consumer = (*Consumer)(nil)
