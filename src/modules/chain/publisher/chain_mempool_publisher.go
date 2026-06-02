package publisher

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	chainpubsub "github.com/alexkalak/whatever-system/src/modules/chain/pubsub"
	chainservice "github.com/alexkalak/whatever-system/src/modules/chain/service"
	"github.com/alexkalak/whatever-system/src/modules/pubsub"
)

type ChainMempoolPublisher struct {
	publisher         pubsub.Publisher
	hashesTopic       string
	transactionsTopic string
}

func NewChainMempoolPublisher(publisher pubsub.Publisher, topics ...string) *ChainMempoolPublisher {
	hashesTopic := chainpubsub.ChainMempoolHashesTopic
	transactionsTopic := chainpubsub.ChainMempoolEventsTopic
	if len(topics) > 0 && topics[0] != "" {
		transactionsTopic = topics[0]
	}
	if len(topics) > 1 && topics[1] != "" {
		hashesTopic = topics[1]
	}

	return &ChainMempoolPublisher{publisher: publisher, hashesTopic: hashesTopic, transactionsTopic: transactionsTopic}
}

func (p *ChainMempoolPublisher) PublishHash(ctx context.Context, event chainservice.ChainMempoolHashChannelEntity) error {
	message := chainpubsub.NewChainMempoolHashMessage(event)

	value, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return p.publisher.Publish(ctx, pubsub.Message{
		Topic: p.hashesTopic,
		Key:   []byte(event.TxHash),
		Value: value,
		Headers: map[string]string{
			"chainId": strconv.FormatUint(uint64(event.ChainID), 10),
			"txHash":  event.TxHash,
			"ts":      event.TS.Format(time.RFC3339Nano),
		},
	})
}

func (p *ChainMempoolPublisher) Publish(ctx context.Context, event chainservice.ChainMempoolEventChannelEntity) error {
	message := chainpubsub.NewChainMempoolEventMessage(event)

	value, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return p.publisher.Publish(ctx, pubsub.Message{
		Topic: p.transactionsTopic,
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
