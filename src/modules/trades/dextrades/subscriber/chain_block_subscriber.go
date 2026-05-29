package subscriber

import (
	"context"
	"encoding/json"

	chainpubsub "github.com/alexkalak/whatever-system/src/modules/chain/pubsub"
	"github.com/alexkalak/whatever-system/src/modules/pubsub"
)

type ChainBlockHandler interface {
	HandleChainBlock(ctx context.Context, block chainpubsub.ChainBlockMessage) error
}

type ChainBlockHandlerFunc func(ctx context.Context, block chainpubsub.ChainBlockMessage) error

func (f ChainBlockHandlerFunc) HandleChainBlock(ctx context.Context, block chainpubsub.ChainBlockMessage) error {
	return f(ctx, block)
}

type ChainBlockSubscriber struct {
	consumer pubsub.Consumer
	handler  ChainBlockHandler
}

func NewChainBlockSubscriber(consumer pubsub.Consumer, handler ChainBlockHandler) *ChainBlockSubscriber {
	return &ChainBlockSubscriber{consumer: consumer, handler: handler}
}

func (s *ChainBlockSubscriber) Start(ctx context.Context) error {
	return s.consumer.Subscribe(ctx, func(ctx context.Context, msg pubsub.Message) error {
		var block chainpubsub.ChainBlockMessage
		if err := json.Unmarshal(msg.Value, &block); err != nil {
			return err
		}
		if err := s.handler.HandleChainBlock(ctx, block); err != nil {
			return err
		}
		return msg.Ack(ctx)
	})
}
