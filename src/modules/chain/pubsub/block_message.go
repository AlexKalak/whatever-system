package pubsub

import (
	"encoding/json"
	"time"

	chainentities "github.com/alexkalak/whatever-system/src/modules/chain/entities"
	chainservice "github.com/alexkalak/whatever-system/src/modules/chain/service"
)

type ChainBlockMessage struct {
	TS              time.Time           `json:"ts"`
	ChainID         uint                `json:"chainId"`
	BlockNumber     uint64              `json:"blockNumber"`
	BlockHash       string              `json:"blockHash"`
	ParentBlockHash string              `json:"parentBlockHash"`
	Events          []ChainEventMessage `json:"events"`
}

type ChainEventMessage struct {
	Type        chainentities.ChainEventType `json:"type"`
	BlockNumber uint64                       `json:"blockNumber"`
	Address     string                       `json:"address"`
	TxHash      string                       `json:"txHash"`
	Data        json.RawMessage              `json:"data"`
}

func NewChainBlockMessage(block chainservice.ChainBlockChannelEntity) (ChainBlockMessage, error) {
	events := make([]ChainEventMessage, 0, len(block.Events))
	for _, event := range block.Events {
		data, err := json.Marshal(event.Data)
		if err != nil {
			return ChainBlockMessage{}, err
		}

		events = append(events, ChainEventMessage{
			Type:        event.Type,
			BlockNumber: event.BlockNumber,
			Address:     event.Address,
			TxHash:      event.TxHash,
			Data:        data,
		})
	}

	return ChainBlockMessage{
		TS:              block.TS,
		ChainID:         block.ChainID,
		BlockNumber:     block.BlockNumber,
		BlockHash:       block.BlockHash,
		ParentBlockHash: block.ParentBlockHash,
		Events:          events,
	}, nil
}
