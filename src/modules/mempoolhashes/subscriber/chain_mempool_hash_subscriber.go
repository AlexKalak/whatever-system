package subscriber

import (
	"context"
	"encoding/json"

	chainpubsub "github.com/alexkalak/whatever-system/src/modules/chain/pubsub"
	"github.com/alexkalak/whatever-system/src/modules/pubsub"
)

type ChainMempoolHashHandler interface {
	HandleChainMempoolHash(ctx context.Context, hash chainpubsub.ChainMempoolHashMessage) error
}

type ChainMempoolHashHandlerFunc func(ctx context.Context, hash chainpubsub.ChainMempoolHashMessage) error

func (f ChainMempoolHashHandlerFunc) HandleChainMempoolHash(ctx context.Context, hash chainpubsub.ChainMempoolHashMessage) error {
	return f(ctx, hash)
}

type ChainMempoolHashSubscriber struct {
	consumer pubsub.Consumer
	handler  ChainMempoolHashHandler
}

func NewChainMempoolHashSubscriber(consumer pubsub.Consumer, handler ChainMempoolHashHandler) *ChainMempoolHashSubscriber {
	return &ChainMempoolHashSubscriber{consumer: consumer, handler: handler}
}

func (s *ChainMempoolHashSubscriber) Start(ctx context.Context) error {
	return s.consumer.Subscribe(ctx, func(ctx context.Context, msg pubsub.Message) error {
		var hash chainpubsub.ChainMempoolHashMessage
		if err := json.Unmarshal(msg.Value, &hash); err != nil {
			return err
		}
		if err := s.handler.HandleChainMempoolHash(ctx, hash); err != nil {
			return err
		}
		return msg.Ack(ctx)
	})
}
