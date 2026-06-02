package publisher

import (
	"context"
	"encoding/json"
	"strconv"

	dexactionsprocessor "github.com/alexkalak/whatever-system/src/modules/chainblocks/processor/dexactionsprocessor"
	"github.com/alexkalak/whatever-system/src/modules/pubsub"
	actionentities "github.com/alexkalak/whatever-system/src/modules/trades/dexactions/entities"
	actionpubsub "github.com/alexkalak/whatever-system/src/modules/trades/dexactions/pubsub"
)

type DexActionPublisher struct {
	publisher pubsub.Publisher
	topic     string
}

func NewDexActionPublisher(publisher pubsub.Publisher, topic ...string) *DexActionPublisher {
	t := actionpubsub.DexActionsCreatedTopic
	if len(topic) > 0 && topic[0] != "" {
		t = topic[0]
	}
	return &DexActionPublisher{publisher: publisher, topic: t}
}
func (p *DexActionPublisher) Publish(ctx context.Context, action dexactionsprocessor.CreatedDexAction) error {
	message, err := newDexActionMessage(action)
	if err != nil {
		return err
	}
	value, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return p.publisher.Publish(ctx, pubsub.Message{Topic: p.topic, Key: []byte(strconv.FormatUint(message.ChainID, 10) + ":" + message.DexAddress), Value: value, Headers: map[string]string{"chainId": strconv.FormatUint(message.ChainID, 10), "dexAddress": message.DexAddress, "blockNumber": strconv.FormatUint(message.BlockNumber, 10), "indexInBlock": strconv.FormatUint(message.IndexInBlock, 10), "indexInTx": strconv.FormatUint(message.IndexInTx, 10), "version": message.Version, "actionType": message.ActionType}})
}
func newDexActionMessage(created dexactionsprocessor.CreatedDexAction) (actionpubsub.DexActionMessage, error) {
	payload, err := json.Marshal(created.Action)
	if err != nil {
		return actionpubsub.DexActionMessage{}, err
	}
	message := actionpubsub.DexActionMessage{Version: created.Version, Action: payload}
	switch action := created.Action.(type) {
	case actionentities.DexActionUniswapV2:
		setV2(&message, action)
	case *actionentities.DexActionUniswapV2:
		setV2(&message, *action)
	case actionentities.DexActionUniswapV3:
		setV3(&message, action)
	case *actionentities.DexActionUniswapV3:
		setV3(&message, *action)
	}
	return message, nil
}
func setV2(message *actionpubsub.DexActionMessage, action actionentities.DexActionUniswapV2) {
	message.ChainID = action.ChainID
	message.DexAddress = action.DexAddress
	message.ActionType = action.ActionType
	message.BlockNumber = action.BlockNumber
	message.IndexInBlock = action.IndexInBlock
	message.IndexInTx = action.IndexInTx
	message.PoolAddress = action.PoolAddress
}
func setV3(message *actionpubsub.DexActionMessage, action actionentities.DexActionUniswapV3) {
	message.ChainID = action.ChainID
	message.DexAddress = action.DexAddress
	message.ActionType = action.ActionType
	message.BlockNumber = action.BlockNumber
	message.IndexInBlock = action.IndexInBlock
	message.IndexInTx = action.IndexInTx
	message.PoolAddress = action.PoolAddress
}
