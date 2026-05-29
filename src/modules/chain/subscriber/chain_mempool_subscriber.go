package subscriber

import (
	"context"
	"encoding/json"

	chainpubsub "github.com/alexkalak/whatever-system/src/modules/chain/pubsub"
	"github.com/alexkalak/whatever-system/src/modules/pubsub"
)

type ChainMempoolHandler interface {
	HandleChainMempoolTransaction(ctx context.Context, tx chainpubsub.ChainMempoolEventMessage) error
}

type ChainMempoolHandlerFunc func(ctx context.Context, tx chainpubsub.ChainMempoolEventMessage) error

func (f ChainMempoolHandlerFunc) HandleChainMempoolTransaction(ctx context.Context, tx chainpubsub.ChainMempoolEventMessage) error {
	return f(ctx, tx)
}

type ChainMempoolSubscriber struct {
	consumer pubsub.Consumer
	handler  ChainMempoolHandler
}

func NewChainMempoolSubscriber(consumer pubsub.Consumer, handler ChainMempoolHandler) *ChainMempoolSubscriber {
	return &ChainMempoolSubscriber{consumer: consumer, handler: handler}
}

func (s *ChainMempoolSubscriber) Start(ctx context.Context) error {
	return s.consumer.Subscribe(ctx, func(ctx context.Context, msg pubsub.Message) error {
		var tx chainpubsub.ChainMempoolEventMessage
		if err := json.Unmarshal(msg.Value, &tx); err != nil {
			return err
		}
		if err := s.handler.HandleChainMempoolTransaction(ctx, tx); err != nil {
			return err
		}
		return msg.Ack(ctx)
	})
}
