package publisher

import (
	"context"
	"encoding/json"
	"strconv"

	chainpubsub "github.com/alexkalak/whatever-system/src/modules/chain/pubsub"
	chainservice "github.com/alexkalak/whatever-system/src/modules/chain/service"
	"github.com/alexkalak/whatever-system/src/modules/pubsub"
)

type ChainBlockPublisher struct {
	publisher pubsub.Publisher
	topic     string
}

func NewChainBlockPublisher(publisher pubsub.Publisher, topic ...string) *ChainBlockPublisher {
	resolvedTopic := chainpubsub.ChainBlocksTopic
	if len(topic) > 0 && topic[0] != "" {
		resolvedTopic = topic[0]
	}

	return &ChainBlockPublisher{publisher: publisher, topic: resolvedTopic}
}

func (p *ChainBlockPublisher) Publish(ctx context.Context, block chainservice.ChainBlockChannelEntity) error {
	message, err := chainpubsub.NewChainBlockMessage(block)
	if err != nil {
		return err
	}

	value, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return p.publisher.Publish(ctx, pubsub.Message{
		Topic: p.topic,
		Key:   []byte(strconv.FormatUint(uint64(block.ChainID), 10)),
		Value: value,
		Headers: map[string]string{
			"chainId":     strconv.FormatUint(uint64(block.ChainID), 10),
			"blockNumber": strconv.FormatUint(block.BlockNumber, 10),
			"blockHash":   block.BlockHash,
		},
	})
}
