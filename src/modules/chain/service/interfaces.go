package service

import (
	"context"
	"time"

	chainentities "github.com/alexkalak/whatever-system/src/modules/chain/entities"
	"github.com/ethereum/go-ethereum/common"
)

type ChainLogsListenerConfig struct {
	ChainID  uint
	WsRPCURL string
}

type ChainEventChannelEntity struct {
	TS      time.Time
	ChainID uint
	Event   chainentities.ChainEvent
}

type ChainLogsListener interface {
	Start(ctx context.Context, fromBlock uint64, addresses []common.Address) (<-chan ChainEventChannelEntity, error)
}

type TokenInfo struct {
	ChainID  uint
	Address  string
	Symbol   string
	Name     string
	Decimals uint8
}

type ChainDataService interface {
	GetTokenInfo(ctx context.Context, chainID uint, tokenAddress string) (TokenInfo, error)
}
