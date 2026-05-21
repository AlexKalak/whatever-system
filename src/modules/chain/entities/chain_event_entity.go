package entities

type ChainEventType = string

const (
	ChainSwapV3Event ChainEventType = "CHAIN_SWAPV3_EVENT"
	ChainMintV3Event ChainEventType = "CHAIN_MINTV3_EVENT"
	ChainBurnV3Event ChainEventType = "CHAIN_BURNV3_EVENT"
	ChainSwapV2Event ChainEventType = "CHAIN_SWAPV2_EVENT"
	ChainSyncV2Event ChainEventType = "CHAIN_SYNCV2_EVENT"
)

type ChainEventData interface {
	isChainEventData()
}

type ChainEvent struct {
	Type        ChainEventType `json:"type"`
	BlockNumber uint64         `json:"block_number"`
	Address     string         `json:"address"`
	TxHash      string         `json:"tx_hash"`
	Data        ChainEventData `json:"data"`
}
