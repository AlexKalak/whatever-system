package publisher

import (
	"context"
	"encoding/json"
	"strconv"

	chainpubsub "github.com/alexkalak/whatever-system/src/modules/chain/pubsub"
	chainservice "github.com/alexkalak/whatever-system/src/modules/chain/service"
	"github.com/alexkalak/whatever-system/src/modules/pubsub"
)

type ChainMempoolPublisher struct {
	publisher pubsub.Publisher
	topic     string
}

func NewChainMempoolPublisher(publisher pubsub.Publisher, topic ...string) *ChainMempoolPublisher {
	resolvedTopic := chainpubsub.ChainMempoolEventsTopic
	if len(topic) > 0 && topic[0] != "" {
		resolvedTopic = topic[0]
	}

	return &ChainMempoolPublisher{publisher: publisher, topic: resolvedTopic}
}

func (p *ChainMempoolPublisher) Publish(ctx context.Context, event chainservice.ChainMempoolEventChannelEntity) error {
	message := chainpubsub.NewChainMempoolEventMessage(event)

	value, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return p.publisher.Publish(ctx, pubsub.Message{
		Topic: p.topic,
		Key:   []byte(event.TxHash),
		Value: value,
		Headers: map[string]string{
			"chainId":   strconv.FormatUint(uint64(event.ChainID), 10),
			"txHash":    event.TxHash,
			"from":      event.From,
			"to":        event.To,
			"method":    event.CallData.Method,
			"signature": event.CallData.Signature,
		},
	})
}
