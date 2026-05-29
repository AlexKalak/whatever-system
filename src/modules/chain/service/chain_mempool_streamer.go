package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type chainMempoolStreamer struct {
	parser  chainMempoolCallDataParser
	client  *ethclient.Client
	chainID uint
}

func NewChainMempoolStreamer(ctx context.Context, config ChainMempoolStreamerConfig) (ChainMempoolStreamer, error) {
	client, err := ethclient.DialContext(ctx, config.WsRPCURL)
	if err != nil {
		log.Println("Unable to init ws mempool client:", err)
		return nil, fmt.Errorf("cannot connect to mempool ws client: %w", err)
	}

	parser, err := newChainMempoolCallDataParser()
	if err != nil {
		client.Close()
		return nil, err
	}

	return &chainMempoolStreamer{
		parser:  parser,
		client:  client,
		chainID: config.ChainID,
	}, nil
}

func (s *chainMempoolStreamer) Start(ctx context.Context) (<-chan ChainMempoolEventChannelEntity, error) {
	hashesCh := make(chan common.Hash, 4096)
	sub, err := s.client.Client().EthSubscribe(ctx, hashesCh, "newPendingTransactions")
	if err != nil {
		s.client.Close()
		return nil, err
	}

	outCh := make(chan ChainMempoolEventChannelEntity, 1024)
	go func() {
		defer close(outCh)
		defer sub.Unsubscribe()
		defer s.client.Close()

		if err := s.streamPendingTransactions(ctx, sub, hashesCh, outCh); err != nil {
			log.Println("chain mempool stream stopped:", err)
		}
	}()

	return outCh, nil
}

func (s *chainMempoolStreamer) streamPendingTransactions(
	ctx context.Context,
	sub ethereum.Subscription,
	hashesCh <-chan common.Hash,
	outCh chan<- ChainMempoolEventChannelEntity,
) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-sub.Err():
			return err
		case hash, ok := <-hashesCh:
			if !ok {
				return errors.New("pending transaction hash channel closed")
			}

			event, ok, err := s.decodePendingTransaction(ctx, hash)
			if err != nil {
				log.Printf("pending tx calldata decode failed hash=%s: %v", hash.String(), err)
				continue
			}
			if !ok {
				continue
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case outCh <- event:
				log.Println("New mempool calldata decoded:", event.ChainID, event.TxHash, event.CallData.Method)
			}
		}
	}
}

func (s *chainMempoolStreamer) decodePendingTransaction(ctx context.Context, hash common.Hash) (ChainMempoolEventChannelEntity, bool, error) {
	tx, _, err := s.client.TransactionByHash(ctx, hash)
	if err != nil {
		return ChainMempoolEventChannelEntity{}, false, err
	}
	if tx == nil {
		return ChainMempoolEventChannelEntity{}, false, nil
	}

	callData, ok, err := s.parser.parse(tx.Data())
	if err != nil || !ok {
		return ChainMempoolEventChannelEntity{}, ok, err
	}

	to := ""
	if tx.To() != nil {
		to = tx.To().String()
	}

	from := ""
	if tx.ChainId() != nil && tx.ChainId().Sign() > 0 {
		signer := types.LatestSignerForChainID(tx.ChainId())
		if sender, err := types.Sender(signer, tx); err == nil {
			from = sender.String()
		}
	} else {
		signer := types.LatestSignerForChainID(new(big.Int).SetUint64(uint64(s.chainID)))
		if sender, err := types.Sender(signer, tx); err == nil {
			from = sender.String()
		}
	}

	return ChainMempoolEventChannelEntity{
		TS:       time.Now(),
		ChainID:  s.chainID,
		TxHash:   tx.Hash().String(),
		From:     from,
		To:       to,
		Value:    bigIntToString(tx.Value()),
		CallData: callData,
	}, true, nil
}

func bigIntToString(value *big.Int) string {
	if value == nil {
		return "0"
	}
	return value.String()
}
