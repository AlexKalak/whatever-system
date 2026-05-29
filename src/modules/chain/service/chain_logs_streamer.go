package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/big"
	"time"

	chainentities "github.com/alexkalak/whatever-system/src/modules/chain/entities"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type chainLogsStreamer struct {
	parser         chainLogsParser
	wsLogsClient   *ethclient.Client
	httpLogsClient *ethclient.Client
	chainID        uint
	wsRPCURL       string
	httpRPCURL     string
}

func NewChainLogsStreamer(ctx context.Context, config ChainLogsStreamerConfig) (ChainLogsStreamer, error) {
	logsWsClient, err := ethclient.DialContext(ctx, config.WsRPCURL)
	if err != nil {
		log.Println("Unable to init ws logs client:", err)
		return nil, fmt.Errorf("cannot connect to logs ws client: %w", err)
	}

	logsHTTPClient, err := ethclient.DialContext(ctx, config.HTTPRPCURL)
	if err != nil {
		logsWsClient.Close()
		log.Println("Unable to init http logs client:", err)
		return nil, fmt.Errorf("cannot connect to logs http client: %w", err)
	}

	parser, err := newChainLogsParser()
	if err != nil {
		logsWsClient.Close()
		logsHTTPClient.Close()
		return nil, err
	}

	return &chainLogsStreamer{
		wsRPCURL:       config.WsRPCURL,
		httpRPCURL:     config.HTTPRPCURL,
		chainID:        config.ChainID,
		wsLogsClient:   logsWsClient,
		httpLogsClient: logsHTTPClient,
		parser:         parser,
	}, nil
}

func (s *chainLogsStreamer) Start(ctx context.Context, fromBlock uint64, addresses []common.Address) (<-chan ChainBlockChannelEntity, error) {
	headersCh := make(chan *types.Header, 1024)
	sub, err := s.wsLogsClient.SubscribeNewHead(ctx, headersCh)
	if err != nil {
		return nil, err
	}

	outCh := make(chan ChainBlockChannelEntity, 64)
	go func() {
		defer close(outCh)
		defer sub.Unsubscribe()
		if err := s.backfillAndStreamBlocks(ctx, sub, fromBlock, addresses, headersCh, outCh); err != nil {
			log.Println("chain logs block stream stopped:", err)
		}
	}()

	return outCh, nil
}

func (s *chainLogsStreamer) getSupportedEventSigs() []common.Hash {
	sigs := make([]common.Hash, 0, len(s.parser.uniswapV2Sigs)+len(s.parser.uniswapV3Sigs)+len(s.parser.pancakeswapV3Sigs)+len(s.parser.sushiswapV3Sigs))
	sigs = append(sigs, s.parser.uniswapV2Sigs...)
	sigs = append(sigs, s.parser.uniswapV3Sigs...)
	sigs = append(sigs, s.parser.pancakeswapV3Sigs...)
	sigs = append(sigs, s.parser.sushiswapV3Sigs...)
	return sigs
}

func (s *chainLogsStreamer) backfillAndStreamBlocks(
	ctx context.Context,
	sub ethereum.Subscription,
	fromBlock uint64,
	addresses []common.Address,
	headersCh <-chan *types.Header,
	outCh chan<- ChainBlockChannelEntity,
) error {
	lastProcessedBlock := uint64(0)

	if fromBlock > 0 {
		currentHeader, err := s.httpLogsClient.HeaderByNumber(ctx, nil)
		if err != nil {
			log.Println("Header by number ERR: ", err)
			return err
		}
		if currentHeader != nil && currentHeader.Number != nil {
			currentBlock := currentHeader.Number.Uint64()
			if fromBlock <= currentBlock {
				if err := s.backfillBlocks(ctx, fromBlock, currentBlock, addresses, outCh); err != nil {
					log.Println("ERR: ", err)
					return err
				}
				lastProcessedBlock = currentBlock
			}
		}
	}

	return s.streamNewBlocks(ctx, sub, lastProcessedBlock, addresses, headersCh, outCh)
}

func (s *chainLogsStreamer) backfillBlocks(ctx context.Context, fromBlock uint64, toBlock uint64, addresses []common.Address, outCh chan<- ChainBlockChannelEntity) error {
	for blockNumber := fromBlock; blockNumber <= toBlock; blockNumber++ {
		log.Println("Backfilling: ", blockNumber)
		if err := s.emitBlock(ctx, blockNumber, addresses, outCh); err != nil {
			log.Println("Err emiting block: ", blockNumber)
			return err
		}
	}
	return nil
}

func (s *chainLogsStreamer) streamNewBlocks(
	ctx context.Context,
	sub ethereum.Subscription,
	lastProcessedBlock uint64,
	addresses []common.Address,
	headersCh <-chan *types.Header,
	outCh chan<- ChainBlockChannelEntity,
) error {
	for {
		select {
		case <-ctx.Done():
			log.Println("Context done")
			return ctx.Err()
		case err := <-sub.Err():
			log.Println("Sub err: ", err)
			return err
		case header, ok := <-headersCh:
			if !ok {
				log.Println("Sub header not ok")
				return errors.New("headers channel closed")
			}
			if header == nil || header.Number == nil {
				continue
			}

			blockNumber := header.Number.Uint64()
			if blockNumber <= lastProcessedBlock {
				continue
			}

			if err := s.emitBlock(ctx, blockNumber, addresses, outCh); err != nil {
				log.Println("Err emiting ws block", err)
				return err
			}
			lastProcessedBlock = blockNumber
		}
	}
}

func (s *chainLogsStreamer) emitBlock(ctx context.Context, blockNumber uint64, addresses []common.Address, outCh chan<- ChainBlockChannelEntity) error {
	block, err := s.loadBlockEvents(ctx, blockNumber, addresses)
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case outCh <- block:
		log.Println("New block from ws emited: ", block.ChainID, block.BlockHash, block.BlockNumber)
		return nil
	}
}

func (s *chainLogsStreamer) loadBlockEvents(ctx context.Context, blockNumber uint64, addresses []common.Address) (ChainBlockChannelEntity, error) {
	block, err := s.httpLogsClient.BlockByNumber(ctx, new(big.Int).SetUint64(blockNumber))
	if err != nil {
		return ChainBlockChannelEntity{}, err
	}
	if block == nil {
		return ChainBlockChannelEntity{}, fmt.Errorf("block %d not found", blockNumber)
	}

	logs, err := s.httpLogsClient.FilterLogs(ctx, ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(blockNumber),
		ToBlock:   new(big.Int).SetUint64(blockNumber),
		Addresses: addresses,
		Topics:    [][]common.Hash{s.getSupportedEventSigs()},
	})
	if err != nil {
		return ChainBlockChannelEntity{}, err
	}

	events := make([]chainentities.ChainEvent, 0, len(logs))
	for _, lg := range logs {
		event, ok := s.parseLog(lg)
		if ok {
			events = append(events, event)
		}
	}

	return ChainBlockChannelEntity{
		TS:              time.Now(),
		ChainID:         s.chainID,
		BlockNumber:     blockNumber,
		BlockHash:       block.Hash().String(),
		ParentBlockHash: block.ParentHash().String(),
		Events:          events,
	}, nil
}

func (s *chainLogsStreamer) parseLog(lg types.Log) (chainentities.ChainEvent, bool) {
	eventType, data, err := s.parser.parse(lg)
	if err != nil {
		return chainentities.ChainEvent{}, false
	}

	return chainentities.ChainEvent{
		Type:        eventType,
		BlockNumber: lg.BlockNumber,
		Address:     lg.Address.String(),
		TxHash:      lg.TxHash.String(),
		Data:        data,
	}, true
}
