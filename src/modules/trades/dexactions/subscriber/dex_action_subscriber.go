package subscriber

import (
	"context"
	"encoding/json"

	"github.com/alexkalak/whatever-system/src/modules/pubsub"
	actionpubsub "github.com/alexkalak/whatever-system/src/modules/trades/dexactions/pubsub"
)

type DexActionHandler interface {
	HandleDexAction(ctx context.Context, action actionpubsub.DexActionMessage) error
}
type DexActionHandlerFunc func(ctx context.Context, action actionpubsub.DexActionMessage) error

func (f DexActionHandlerFunc) HandleDexAction(ctx context.Context, action actionpubsub.DexActionMessage) error {
	return f(ctx, action)
}

type DexActionSubscriber struct {
	consumer pubsub.Consumer
	handler  DexActionHandler
}

func NewDexActionSubscriber(consumer pubsub.Consumer, handler DexActionHandler) *DexActionSubscriber {
	return &DexActionSubscriber{consumer: consumer, handler: handler}
}
func (s *DexActionSubscriber) Start(ctx context.Context) error {
	return s.consumer.Subscribe(ctx, func(ctx context.Context, msg pubsub.Message) error {
		var action actionpubsub.DexActionMessage
		if err := json.Unmarshal(msg.Value, &action); err != nil {
			return err
		}
		if err := s.handler.HandleDexAction(ctx, action); err != nil {
			return err
		}
		return msg.Ack(ctx)
	})
}
