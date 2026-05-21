package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	chainentities "github.com/alexkalak/whatever-system/src/modules/chain/entities"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type chainLogsListener struct {
	handler      chainLogsParser
	wsLogsClient *ethclient.Client
	chainID      uint
	wsRPCURL     string
}

func NewChainLogsListener(ctx context.Context, config ChainLogsListenerConfig) (ChainLogsListener, error) {
	logsWsClient, err := ethclient.DialContext(ctx, config.WsRPCURL)
	if err != nil {
		log.Println("Unable to init ws logs client:", err)
		return nil, fmt.Errorf("cannot connect to logs ws client: %w", err)
	}
	chainLogsHandler, err := newChainLogsParser()
	if err != nil {
		return nil, err
	}
	return &chainLogsListener{wsRPCURL: config.WsRPCURL, chainID: config.ChainID, wsLogsClient: logsWsClient, handler: chainLogsHandler}, nil
}

func (c *chainLogsListener) Start(ctx context.Context, fromBlock uint64, addresses []common.Address) (<-chan ChainEventChannelEntity, error) {
	sigs := c.getSupportedEventSigs()
	query := ethereum.FilterQuery{Addresses: addresses, Topics: [][]common.Hash{sigs}}
	logsCh := make(chan types.Log, 1024)
	sub, err := c.wsLogsClient.SubscribeFilterLogs(ctx, query, logsCh)
	if err != nil {
		return nil, err
	}
	outCh := make(chan ChainEventChannelEntity, 1024)
	go func() {
		defer close(outCh)
		defer sub.Unsubscribe()
		if err := c.listenNewLogs(ctx, sub, fromBlock, logsCh, outCh); err != nil {
			log.Println("listen new logs stopped:", err)
		}
	}()
	return outCh, nil
}

func (c *chainLogsListener) getSupportedEventSigs() []common.Hash {
	sigs := make([]common.Hash, 0, len(c.handler.uniswapV2Sigs)+len(c.handler.uniswapV3Sigs)+len(c.handler.pancakeswapV3Sigs)+len(c.handler.sushiswapV3Sigs))
	sigs = append(sigs, c.handler.uniswapV2Sigs...)
	sigs = append(sigs, c.handler.uniswapV3Sigs...)
	sigs = append(sigs, c.handler.pancakeswapV3Sigs...)
	sigs = append(sigs, c.handler.sushiswapV3Sigs...)
	return sigs
}

func (c *chainLogsListener) listenNewLogs(ctx context.Context, sub ethereum.Subscription, fromBlock uint64, logsCh <-chan types.Log, outCh chan<- ChainEventChannelEntity) error {
	for {
		select {
		case <-ctx.Done():
			return errors.New("rpc service stopped because of ctx done")
		case err := <-sub.Err():
			return err
		case lg, ok := <-logsCh:
			if !ok {
				return errors.New("logs channel closed")
			}
			if fromBlock > lg.BlockNumber {
				continue
			}
			c.processNewLog(ctx, lg, outCh)
		}
	}
}

func (c *chainLogsListener) processNewLog(ctx context.Context, lg types.Log, outCh chan<- ChainEventChannelEntity) {
	eventType, event, err := c.handler.parse(lg)
	if err != nil {
		return
	}
	chanEntity := ChainEventChannelEntity{TS: time.Now(), ChainID: c.chainID, Event: chainentities.ChainEvent{Type: eventType, BlockNumber: lg.BlockNumber, Address: lg.Address.String(), TxHash: lg.TxHash.String(), Data: event}}
	select {
	case <-ctx.Done():
		return
	case outCh <- chanEntity:
	}
}
