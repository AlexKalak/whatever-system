package pubsub

import "encoding/json"

type DexActionMessage struct {
	Version      string          `json:"version"`
	ChainID      uint64          `json:"chainId"`
	DexAddress   string          `json:"dexAddress"`
	ActionType   string          `json:"actionType"`
	BlockNumber  uint64          `json:"blockNumber"`
	IndexInBlock uint64          `json:"indexInBlock"`
	IndexInTx    uint64          `json:"indexInTx"`
	PoolAddress  string          `json:"poolAddress"`
	Action       json.RawMessage `json:"action"`
}
