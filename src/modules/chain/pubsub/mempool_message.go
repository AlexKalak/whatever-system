package pubsub

import (
	"time"

	chainservice "github.com/alexkalak/whatever-system/src/modules/chain/service"
)

type ChainMempoolCallDataMessage struct {
	Method    string         `json:"method"`
	Signature string         `json:"signature"`
	Selector  string         `json:"selector"`
	Args      map[string]any `json:"args"`
}

type ChainMempoolEventMessage struct {
	TS       time.Time                   `json:"ts"`
	ChainID  uint                        `json:"chainId"`
	TxHash   string                      `json:"txHash"`
	From     string                      `json:"from"`
	To       string                      `json:"to"`
	Value    string                      `json:"value"`
	CallData ChainMempoolCallDataMessage `json:"callData"`
}

func NewChainMempoolEventMessage(event chainservice.ChainMempoolEventChannelEntity) ChainMempoolEventMessage {
	return ChainMempoolEventMessage{
		TS:      event.TS,
		ChainID: event.ChainID,
		TxHash:  event.TxHash,
		From:    event.From,
		To:      event.To,
		Value:   event.Value,
		CallData: ChainMempoolCallDataMessage{
			Method:    event.CallData.Method,
			Signature: event.CallData.Signature,
			Selector:  event.CallData.Selector,
			Args:      event.CallData.Args,
		},
	}
}
