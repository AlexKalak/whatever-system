package chain

import "github.com/alexkalak/whatever-system/src/modules/eventbus"

var (
	SwapV3Topic  = eventbus.NewTopic[SwapV3Payload]("chain.chain_swapv3_event")
	UnknownTopic = eventbus.NewTopic[UnknownPayload]("chain.chain_unknown_event")
)

type SwapV3Payload struct {
	ChainID     uint   `json:"chain_id"`
	Dex         string `json:"dex"`
	BlockNumber uint64 `json:"block_number"`
	PoolAddress string `json:"pool_address"`
	TxHash      string `json:"tx_hash"`
	Sender      string `json:"sender"`
	Recipient   string `json:"recipient"`
	Amount0     string `json:"amount0"`
	Amount1     string `json:"amount1"`
}

type UnknownPayload struct {
	EventType   string `json:"event_type"`
	BlockNumber uint64 `json:"block_number"`
	Address     string `json:"address"`
	TxHash      string `json:"tx_hash"`
}
