package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/alexkalak/whatever-system/src/shared/chaindata/abis"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type evmChainDataService struct {
	clients  map[uint]*ethclient.Client
	erc20ABI abi.ABI
}

func NewEVMChainDataService(rpcByChainID map[uint]string) (ChainDataService, error) {
	erc20ABI, err := abi.JSON(strings.NewReader(abis.ERC20MetadataABIString))
	if err != nil {
		return nil, err
	}
	clients := make(map[uint]*ethclient.Client, len(rpcByChainID))
	for chainID, rpcURL := range rpcByChainID {
		client, err := ethclient.Dial(rpcURL)
		if err != nil {
			return nil, fmt.Errorf("dial chain %d: %w", chainID, err)
		}
		clients[chainID] = client
	}
	return &evmChainDataService{clients: clients, erc20ABI: erc20ABI}, nil
}

func (s *evmChainDataService) GetTokenInfo(ctx context.Context, chainID uint, tokenAddress string) (TokenInfo, error) {
	client, ok := s.clients[chainID]
	if !ok {
		return TokenInfo{}, fmt.Errorf("no rpc client for chain %d", chainID)
	}

	addr := common.HexToAddress(tokenAddress)

	nameData, err := s.erc20ABI.Pack("name")
	if err != nil {
		return TokenInfo{}, err
	}
	symbolData, err := s.erc20ABI.Pack("symbol")
	if err != nil {
		return TokenInfo{}, err
	}
	decimalsData, err := s.erc20ABI.Pack("decimals")
	if err != nil {
		return TokenInfo{}, err
	}

	var nameRaw hexutil.Bytes
	var symbolRaw hexutil.Bytes
	var decimalsRaw hexutil.Bytes

	batch := []rpc.BatchElem{
		{
			Method: "eth_call",
			Args:   []any{map[string]any{"to": addr.Hex(), "data": hexutil.Encode(nameData)}, "latest"},
			Result: &nameRaw,
		},
		{
			Method: "eth_call",
			Args:   []any{map[string]any{"to": addr.Hex(), "data": hexutil.Encode(symbolData)}, "latest"},
			Result: &symbolRaw,
		},
		{
			Method: "eth_call",
			Args:   []any{map[string]any{"to": addr.Hex(), "data": hexutil.Encode(decimalsData)}, "latest"},
			Result: &decimalsRaw,
		},
	}

	if err := client.Client().BatchCallContext(ctx, batch); err != nil {
		return TokenInfo{}, err
	}
	for _, elem := range batch {
		if elem.Error != nil {
			return TokenInfo{}, elem.Error
		}
	}

	nameVals, err := s.erc20ABI.Unpack("name", nameRaw)
	if err != nil || len(nameVals) == 0 {
		return TokenInfo{}, fmt.Errorf("unpack name failed: %w", err)
	}
	name, ok := nameVals[0].(string)
	if !ok {
		return TokenInfo{}, fmt.Errorf("unexpected name return type")
	}

	symbolVals, err := s.erc20ABI.Unpack("symbol", symbolRaw)
	if err != nil || len(symbolVals) == 0 {
		return TokenInfo{}, fmt.Errorf("unpack symbol failed: %w", err)
	}
	symbol, ok := symbolVals[0].(string)
	if !ok {
		return TokenInfo{}, fmt.Errorf("unexpected symbol return type")
	}

	decimalsVals, err := s.erc20ABI.Unpack("decimals", decimalsRaw)
	if err != nil || len(decimalsVals) == 0 {
		return TokenInfo{}, fmt.Errorf("unpack decimals failed: %w", err)
	}
	decimals, ok := decimalsVals[0].(uint8)
	if !ok {
		return TokenInfo{}, fmt.Errorf("unexpected decimals return type")
	}

	return TokenInfo{ChainID: chainID, Address: strings.ToLower(addr.Hex()), Name: name, Symbol: symbol, Decimals: decimals}, nil
}
