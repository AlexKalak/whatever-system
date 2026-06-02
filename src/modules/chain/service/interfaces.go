package service

import (
	"context"
	"time"

	chainentities "github.com/alexkalak/whatever-system/src/modules/chain/entities"
	"github.com/ethereum/go-ethereum/common"
)

type TokenInfo struct {
	ChainID  uint
	Address  string
	Symbol   string
	Name     string
	Decimals uint8
}

type UniswapV2PairInfo struct {
	Token0Address string
	Token1Address string
	Reserve0      string
	Reserve1      string
	FeeTier       uint
}

type UniswapV3PoolInfo struct {
	Token0Address string
	Token1Address string
	FeeTier       uint
	SqrtPriceX96  string
	Liquidity     string
	Tick          int64
	TickSpacing   int64
}

type ChainLogsStreamerConfig struct {
	ChainID    uint
	WsRPCURL   string
	HTTPRPCURL string
}

type ChainMempoolStreamerConfig struct {
	ChainID            uint
	WsRPCURL           string
	PendingTxLogPath   string
	SubscriptionMethod string
	SubscriptionParams []any
}

type ChainDataServiceConfig struct {
	RPCByChainID       map[uint]string
	MulticallByChainID map[uint]string
}

type ChainBlockChannelEntity struct {
	TS              time.Time                  `json:"ts"`
	ChainID         uint                       `json:"chainId"`
	BlockNumber     uint64                     `json:"blockNumber"`
	BlockHash       string                     `json:"blockHash"`
	ParentBlockHash string                     `json:"parentBlockHash"`
	Events          []chainentities.ChainEvent `json:"events"`
}

type ChainLogsStreamer interface {
	Start(ctx context.Context, fromBlock uint64, addresses []common.Address) (<-chan ChainBlockChannelEntity, error)
}

type ChainMempoolCallData struct {
	Method    string         `json:"method"`
	Signature string         `json:"signature"`
	Selector  string         `json:"selector"`
	Args      map[string]any `json:"args"`
}

type ChainMempoolHashChannelEntity struct {
	TS      time.Time `json:"ts"`
	ChainID uint      `json:"chainId"`
	TxHash  string    `json:"txHash"`
}

type ChainMempoolEventChannelEntity struct {
	TS       time.Time            `json:"ts"`
	ChainID  uint                 `json:"chainId"`
	TxHash   string               `json:"txHash"`
	From     string               `json:"from"`
	To       string               `json:"to"`
	Value    string               `json:"value"`
	Input    string               `json:"input"`
	CallData ChainMempoolCallData `json:"callData"`
}

type ChainMempoolStreams struct {
	Hashes       <-chan ChainMempoolHashChannelEntity
	Transactions <-chan ChainMempoolEventChannelEntity
}

type ChainMempoolStreamer interface {
	Start(ctx context.Context) (ChainMempoolStreams, error)
}

type ChainDataService interface {
	GetTokenInfo(ctx context.Context, chainID uint, tokenAddress string) (TokenInfo, error)
	GetPoolFeeTier(ctx context.Context, chainID uint, poolAddress string) (uint, error)
	GetUniswapV2PairInfo(ctx context.Context, chainID uint, pairAddress string) (UniswapV2PairInfo, error)
	GetUniswapV3PoolInfo(ctx context.Context, chainID uint, poolAddress string) (UniswapV3PoolInfo, error)
}
