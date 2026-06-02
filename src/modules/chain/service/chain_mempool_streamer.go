package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alexkalak/whatever-system/src/shared/tools"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type chainMempoolStreamer struct {
	parser             chainMempoolCallDataParser
	client             *ethclient.Client
	chainID            uint
	pendingTxLogPath   string
	subscriptionMethod string
	subscriptionParams []any
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

	subscriptionMethod := config.SubscriptionMethod
	if subscriptionMethod == "" {
		subscriptionMethod = "newPendingTransactions"
	}

	subscriptionParams := config.SubscriptionParams
	if subscriptionMethod == "newPendingTransactions" && subscriptionParams == nil {
		subscriptionParams = []any{true}
	}

	return &chainMempoolStreamer{
		parser:             parser,
		client:             client,
		chainID:            config.ChainID,
		pendingTxLogPath:   config.PendingTxLogPath,
		subscriptionMethod: subscriptionMethod,
		subscriptionParams: subscriptionParams,
	}, nil
}

func (s *chainMempoolStreamer) Start(ctx context.Context) (ChainMempoolStreams, error) {
	pendingTxCh := make(chan json.RawMessage, 4096)
	subscribeArgs := append([]any{s.subscriptionMethod}, s.subscriptionParams...)
	sub, err := s.client.Client().EthSubscribe(ctx, pendingTxCh, subscribeArgs...)
	if err != nil {
		s.client.Close()
		return ChainMempoolStreams{}, err
	}

	logFile, err := s.openPendingTxLog()
	if err != nil {
		sub.Unsubscribe()
		s.client.Close()
		return ChainMempoolStreams{}, err
	}

	hashesCh := make(chan ChainMempoolHashChannelEntity, 4096)
	transactionsCh := make(chan ChainMempoolEventChannelEntity, 1024)
	go func() {
		defer close(hashesCh)
		defer close(transactionsCh)
		defer sub.Unsubscribe()
		defer s.client.Close()
		if logFile != nil {
			defer logFile.Close()
		}

		if err := s.streamPendingTransactions(ctx, sub, pendingTxCh, hashesCh, transactionsCh, logFile); err != nil {
			log.Println("chain mempool stream stopped:", err)
		}
	}()

	return ChainMempoolStreams{Hashes: hashesCh, Transactions: transactionsCh}, nil
}

func (s *chainMempoolStreamer) openPendingTxLog() (*os.File, error) {
	if s.pendingTxLogPath == "" {
		return nil, nil
	}

	if err := os.MkdirAll(filepath.Dir(s.pendingTxLogPath), 0o755); err != nil {
		return nil, err
	}

	return os.OpenFile(s.pendingTxLogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
}

func (s *chainMempoolStreamer) streamPendingTransactions(
	ctx context.Context,
	sub ethereum.Subscription,
	pendingTxCh <-chan json.RawMessage,
	hashesCh chan<- ChainMempoolHashChannelEntity,
	transactionsCh chan<- ChainMempoolEventChannelEntity,
	logFile *os.File,
) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-sub.Err():
			return err
		case rawTx, ok := <-pendingTxCh:
			if !ok {
				return errors.New("pending transaction channel closed")
			}

			if txHash := s.extractPendingTransactionHash(rawTx); txHash != "" {
				hashEvent := ChainMempoolHashChannelEntity{TS: time.Now(), ChainID: s.chainID, TxHash: txHash}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case hashesCh <- hashEvent:
				}
			}

			event, ok, err := s.decodePendingTransaction(ctx, rawTx)
			if err != nil {
				log.Printf("pending tx calldata decode failed: %v", err)
				continue
			}
			if !ok {
				continue
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case transactionsCh <- event:
			}
		}
	}
}

func (s *chainMempoolStreamer) extractPendingTransactionHash(rawTx json.RawMessage) string {
	if isJSONString(rawTx) {
		var hash string
		if err := json.Unmarshal(rawTx, &hash); err == nil {
			return hash
		}
		return ""
	}

	var tx pendingTransactionPayload
	if err := json.Unmarshal(rawTx, &tx); err != nil {
		return ""
	}
	if tx.Hash == (common.Hash{}) {
		return ""
	}
	return tx.Hash.String()
}

type pendingTransactionPayload struct {
	Hash  common.Hash     `json:"hash"`
	From  *common.Address `json:"from"`
	To    *common.Address `json:"to"`
	Value *hexutil.Big    `json:"value"`
	Input hexutil.Bytes   `json:"input"`
}

func (s *chainMempoolStreamer) decodePendingTransaction(ctx context.Context, rawTx json.RawMessage) (ChainMempoolEventChannelEntity, bool, error) {
	fmt.Println("Processing: ", string(rawTx))
	if isJSONString(rawTx) {
		return ChainMempoolEventChannelEntity{}, false, nil
	}

	var tx pendingTransactionPayload
	if err := json.Unmarshal(rawTx, &tx); err != nil {
		return ChainMempoolEventChannelEntity{}, false, err
	}

	input := tx.Input
	if len(input) == 0 {
		return ChainMempoolEventChannelEntity{}, false, nil
	}

	callData, ok, err := s.parser.parse(input)
	if err != nil || !ok {
		log.Printf("pending tx calldata parse failed hash=%s: %v", tx.Hash.String(), err)
		return ChainMempoolEventChannelEntity{}, false, errors.New("unable to parse tx input")
	}

	txHash := tx.Hash.String()
	fmt.Println("Pending ", txHash, tools.GetJSONString(callData))

	from := ""
	if tx.From != nil {
		from = tx.From.String()
	}

	to := ""
	if tx.To != nil {
		to = tx.To.String()
	}

	value := "0"
	if tx.Value != nil {
		value = bigIntToString(tx.Value.ToInt())
	}

	return ChainMempoolEventChannelEntity{
		TS:       time.Now(),
		ChainID:  s.chainID,
		TxHash:   txHash,
		From:     from,
		To:       to,
		Value:    value,
		CallData: callData,
	}, true, nil
}

func (s *chainMempoolStreamer) decodePendingTransactionByHash(ctx context.Context, hash common.Hash) (ChainMempoolEventChannelEntity, bool, error) {
	tx, _, err := s.client.TransactionByHash(ctx, hash)
	if err != nil {
		return ChainMempoolEventChannelEntity{}, false, err
	}
	if tx == nil {
		return ChainMempoolEventChannelEntity{}, false, nil
	}

	if len(tx.Data()) == 0 {
		return ChainMempoolEventChannelEntity{}, false, nil
	}

	callData, _, err := s.parser.parse(tx.Data())
	if err != nil {
		log.Printf("pending tx calldata parse failed hash=%s: %v", tx.Hash().String(), err)
	}

	to := ""
	if tx.To() != nil {
		to = tx.To().String()
	}

	from := ""
	signer := types.LatestSignerForChainID(new(big.Int).SetUint64(uint64(s.chainID)))
	if tx.ChainId() != nil && tx.ChainId().Sign() > 0 {
		signer = types.LatestSignerForChainID(tx.ChainId())
	}
	if sender, err := types.Sender(signer, tx); err == nil {
		from = sender.String()
	}

	return ChainMempoolEventChannelEntity{
		TS:       time.Now(),
		ChainID:  s.chainID,
		TxHash:   tx.Hash().String(),
		From:     from,
		To:       to,
		Value:    bigIntToString(tx.Value()),
		Input:    hexutil.Encode(tx.Data()),
		CallData: callData,
	}, true, nil
}

func isJSONString(raw json.RawMessage) bool {
	return strings.HasPrefix(strings.TrimSpace(string(raw)), "\"")
}

func bigIntToString(value *big.Int) string {
	if value == nil {
		return "0"
	}
	return value.String()
}
