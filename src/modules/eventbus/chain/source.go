package chain

import (
	"context"

	chainservice "github.com/alexkalak/whatever-system/src/modules/chain/service"
	"github.com/alexkalak/whatever-system/src/modules/eventbus"
	"github.com/ethereum/go-ethereum/common"
)

type SourceConfig struct {
	ChainID  uint
	WsRPCURL string
}

func StartSource(
	ctx context.Context,
	sender eventbus.Sender,
	cfg SourceConfig,
	fromBlock uint64,
	addresses []common.Address,
) error {
	listener, err := chainservice.NewChainLogsListener(ctx, chainservice.ChainLogsListenerConfig{
		ChainID:  cfg.ChainID,
		WsRPCURL: cfg.WsRPCURL,
	})
	if err != nil {
		return err
	}

	chainEventsCh, err := listener.Start(ctx, fromBlock, addresses)
	if err != nil {
		return err
	}

	adapter := NewEventsAdapter(sender)
	return adapter.Forward(ctx, chainEventsCh)
}
