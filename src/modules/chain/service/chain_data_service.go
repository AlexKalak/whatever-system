package service

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/alexkalak/whatever-system/src/shared/chaindata/abis"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type evmChainDataService struct {
	clients                     map[uint]*ethclient.Client
	erc20ABI                    abi.ABI
	poolV3ABI                   abi.ABI
	uniswapV2ABI                abi.ABI
	uniswapV3ABI                abi.ABI
	uniswapV3Slot0NoUnlockedABI abi.ABI
	multicall3ABI               abi.ABI
	multicallAddresses          map[uint]common.Address
}

type multicall3Call struct {
	Target       common.Address
	AllowFailure bool
	CallData     []byte
}

type multicall3Result struct {
	Success    bool
	ReturnData []byte
}

func NewEVMChainDataService(rpcByChainID map[uint]string, multicallByChainID ...map[uint]string) (ChainDataService, error) {
	erc20ABI, err := abi.JSON(strings.NewReader(abis.ERC20MetadataABIString))
	if err != nil {
		return nil, err
	}
	poolV3ABI, err := abi.JSON(strings.NewReader(abis.UniswapV3PoolABIString))
	if err != nil {
		return nil, err
	}
	uniswapV2ABI, err := abi.JSON(strings.NewReader(abis.UniswapV2PairABIString))
	if err != nil {
		return nil, err
	}
	uniswapV3ABI, err := abi.JSON(strings.NewReader(abis.UniswapV3PoolABIString))
	if err != nil {
		return nil, err
	}
	uniswapV3Slot0NoUnlockedABI, err := abi.JSON(strings.NewReader(abis.UniswapV3PoolSlot0NoUnlockedABIString))
	if err != nil {
		return nil, err
	}
	multicall3ABI, err := abi.JSON(strings.NewReader(abis.Multicall3ABIString))
	if err != nil {
		return nil, err
	}
	multicallAddresses := make(map[uint]common.Address)
	if len(multicallByChainID) > 0 {
		for chainID, address := range multicallByChainID[0] {
			address = strings.TrimSpace(address)
			if address == "" {
				continue
			}
			multicallAddresses[chainID] = common.HexToAddress(address)
		}
	}

	clients := make(map[uint]*ethclient.Client, len(rpcByChainID))
	for chainID, rpcURL := range rpcByChainID {
		client, err := ethclient.Dial(rpcURL)
		if err != nil {
			return nil, fmt.Errorf("dial chain %d: %w", chainID, err)
		}
		clients[chainID] = client
	}
	return &evmChainDataService{
		clients:                     clients,
		erc20ABI:                    erc20ABI,
		poolV3ABI:                   poolV3ABI,
		uniswapV2ABI:                uniswapV2ABI,
		uniswapV3ABI:                uniswapV3ABI,
		uniswapV3Slot0NoUnlockedABI: uniswapV3Slot0NoUnlockedABI,
		multicall3ABI:               multicall3ABI,
		multicallAddresses:          multicallAddresses,
	}, nil
}

func (s *evmChainDataService) GetUniswapV2PairInfo(ctx context.Context, chainID uint, pairAddress string) (UniswapV2PairInfo, error) {
	log.Println("Getting uniswapv2pairinfo: ", chainID, pairAddress)
	addr := common.HexToAddress(pairAddress)

	token0Data, _ := s.uniswapV2ABI.Pack("token0")
	token1Data, _ := s.uniswapV2ABI.Pack("token1")
	reservesData, _ := s.uniswapV2ABI.Pack("getReserves")

	results, err := s.multicall(ctx, chainID, []multicall3Call{
		{Target: addr, AllowFailure: false, CallData: token0Data},
		{Target: addr, AllowFailure: false, CallData: token1Data},
		{Target: addr, AllowFailure: false, CallData: reservesData},
	})
	if err != nil {
		return UniswapV2PairInfo{}, fmt.Errorf("v2 pair on-chain multicall error chain=%d pair=%s: %w", chainID, pairAddress, err)
	}
	token0Raw := results[0]
	token1Raw := results[1]
	reservesRaw := results[2]

	token0Vals, err := s.uniswapV2ABI.Unpack("token0", token0Raw)
	if err != nil || len(token0Vals) == 0 {
		return UniswapV2PairInfo{}, fmt.Errorf("unpack token0 failed: %w", err)
	}
	token1Vals, err := s.uniswapV2ABI.Unpack("token1", token1Raw)
	if err != nil || len(token1Vals) == 0 {
		return UniswapV2PairInfo{}, fmt.Errorf("unpack token1 failed: %w", err)
	}
	reservesVals, err := s.uniswapV2ABI.Unpack("getReserves", reservesRaw)
	if err != nil || len(reservesVals) < 2 {
		return UniswapV2PairInfo{}, fmt.Errorf("unpack getReserves failed: %w", err)
	}

	token0, ok := token0Vals[0].(common.Address)
	if !ok {
		return UniswapV2PairInfo{}, fmt.Errorf("unexpected token0 type")
	}
	token1, ok := token1Vals[0].(common.Address)
	if !ok {
		return UniswapV2PairInfo{}, fmt.Errorf("unexpected token1 type")
	}
	reserve0, ok := reservesVals[0].(*big.Int)
	if !ok {
		return UniswapV2PairInfo{}, fmt.Errorf("unexpected reserve0 type")
	}
	reserve1, ok := reservesVals[1].(*big.Int)
	if !ok {
		return UniswapV2PairInfo{}, fmt.Errorf("unexpected reserve1 type")
	}

	return UniswapV2PairInfo{
		Token0Address: strings.ToLower(token0.Hex()),
		Token1Address: strings.ToLower(token1.Hex()),
		Reserve0:      reserve0.String(),
		Reserve1:      reserve1.String(),
		FeeTier:       3000,
	}, nil
}

func (s *evmChainDataService) GetUniswapV3PoolInfo(ctx context.Context, chainID uint, poolAddress string) (UniswapV3PoolInfo, error) {
	log.Println("Getting uniswapv3poolinfo: ", chainID, poolAddress)
	addr := common.HexToAddress(poolAddress)

	token0Data, _ := s.uniswapV3ABI.Pack("token0")
	token1Data, _ := s.uniswapV3ABI.Pack("token1")
	feeData, _ := s.uniswapV3ABI.Pack("fee")
	liqData, _ := s.uniswapV3ABI.Pack("liquidity")
	tickSpacingData, _ := s.uniswapV3ABI.Pack("tickSpacing")
	slot0Data, _ := s.uniswapV3ABI.Pack("slot0")

	results, err := s.multicall(ctx, chainID, []multicall3Call{
		{Target: addr, AllowFailure: false, CallData: token0Data},
		{Target: addr, AllowFailure: false, CallData: token1Data},
		{Target: addr, AllowFailure: false, CallData: feeData},
		{Target: addr, AllowFailure: false, CallData: liqData},
		{Target: addr, AllowFailure: false, CallData: tickSpacingData},
		{Target: addr, AllowFailure: false, CallData: slot0Data},
	})
	if err != nil {
		return UniswapV3PoolInfo{}, fmt.Errorf("v3 pool on-chain multicall error chain=%d pool=%s: %w", chainID, poolAddress, err)
	}
	token0Raw := results[0]
	token1Raw := results[1]
	feeRaw := results[2]
	liqRaw := results[3]
	tickSpacingRaw := results[4]
	slot0Raw := results[5]

	token0Vals, err := s.uniswapV3ABI.Unpack("token0", token0Raw)
	if err != nil || len(token0Vals) == 0 {
		return UniswapV3PoolInfo{}, fmt.Errorf("unpack token0 failed: %w", err)
	}
	token1Vals, err := s.uniswapV3ABI.Unpack("token1", token1Raw)
	if err != nil || len(token1Vals) == 0 {
		return UniswapV3PoolInfo{}, fmt.Errorf("unpack token1 failed: %w", err)
	}
	feeVals, err := s.uniswapV3ABI.Unpack("fee", feeRaw)
	if err != nil || len(feeVals) == 0 {
		return UniswapV3PoolInfo{}, fmt.Errorf("unpack fee failed: %w", err)
	}
	liqVals, err := s.uniswapV3ABI.Unpack("liquidity", liqRaw)
	if err != nil || len(liqVals) == 0 {
		return UniswapV3PoolInfo{}, fmt.Errorf("unpack liquidity failed: %w", err)
	}
	tickSpacingVals, err := s.uniswapV3ABI.Unpack("tickSpacing", tickSpacingRaw)
	if err != nil || len(tickSpacingVals) == 0 {
		return UniswapV3PoolInfo{}, fmt.Errorf("unpack tickSpacing failed: %w", err)
	}
	slot0Vals, err := s.unpackV3Slot0(slot0Raw)
	if err != nil || len(slot0Vals) < 2 {
		return UniswapV3PoolInfo{}, fmt.Errorf("unpack slot0 failed: %w", err)
	}

	token0, ok := token0Vals[0].(common.Address)
	if !ok {
		return UniswapV3PoolInfo{}, fmt.Errorf("unexpected token0 type")
	}
	token1, ok := token1Vals[0].(common.Address)
	if !ok {
		return UniswapV3PoolInfo{}, fmt.Errorf("unexpected token1 type")
	}
	feeTier, err := anyToUint(feeVals[0])
	if err != nil {
		return UniswapV3PoolInfo{}, err
	}
	liq, ok := liqVals[0].(*big.Int)
	if !ok {
		return UniswapV3PoolInfo{}, fmt.Errorf("unexpected liquidity type")
	}
	tickSpacing, err := anyToInt64(tickSpacingVals[0])
	if err != nil {
		return UniswapV3PoolInfo{}, err
	}
	sqrtPrice, err := anyToBigInt(slot0Vals[0])
	if err != nil {
		return UniswapV3PoolInfo{}, err
	}
	tick, err := anyToInt64(slot0Vals[1])
	if err != nil {
		return UniswapV3PoolInfo{}, err
	}

	return UniswapV3PoolInfo{
		Token0Address: strings.ToLower(token0.Hex()),
		Token1Address: strings.ToLower(token1.Hex()),
		FeeTier:       feeTier,
		SqrtPriceX96:  sqrtPrice.String(),
		Liquidity:     liq.String(),
		Tick:          tick,
		TickSpacing:   tickSpacing,
	}, nil
}

func ethCallArg(contract common.Address, data []byte) map[string]any {
	return map[string]any{"to": contract.Hex(), "data": hexutil.Encode(data)}
}

func (s *evmChainDataService) unpackV3Slot0(raw []byte) ([]any, error) {
	vals, err := s.uniswapV3ABI.Unpack("slot0", raw)
	if err == nil {
		return vals, nil
	}

	fallbackVals, fallbackErr := s.uniswapV3Slot0NoUnlockedABI.Unpack("slot0", raw)
	if fallbackErr == nil {
		return fallbackVals, nil
	}

	return nil, err
}

func (s *evmChainDataService) multicall(ctx context.Context, chainID uint, calls []multicall3Call) ([][]byte, error) {
	client, ok := s.clients[chainID]
	if !ok {
		return nil, fmt.Errorf("no rpc client for chain %d", chainID)
	}
	multicallAddress, ok := s.multicallAddresses[chainID]
	if !ok {
		return nil, fmt.Errorf("no multicall contract for chain %d", chainID)
	}

	data, err := s.multicall3ABI.Pack("aggregate3", calls)
	if err != nil {
		return nil, err
	}
	out, err := client.CallContract(ctx, ethereum.CallMsg{To: &multicallAddress, Data: data}, nil)
	if err != nil {
		return nil, err
	}

	var results []multicall3Result
	if err := s.multicall3ABI.UnpackIntoInterface(&results, "aggregate3", out); err != nil {
		return nil, err
	}
	if len(results) != len(calls) {
		return nil, fmt.Errorf("unexpected multicall results count: got=%d want=%d", len(results), len(calls))
	}

	returnData := make([][]byte, len(results))
	for i, result := range results {
		if !result.Success {
			return nil, fmt.Errorf("multicall item %d failed", i)
		}
		returnData[i] = result.ReturnData
	}
	return returnData, nil
}

func (s *evmChainDataService) GetPoolFeeTier(ctx context.Context, chainID uint, poolAddress string) (uint, error) {
	client, ok := s.clients[chainID]
	if !ok {
		return 0, fmt.Errorf("no rpc client for chain %d", chainID)
	}

	addr := common.HexToAddress(poolAddress)
	data, err := s.poolV3ABI.Pack("fee")
	if err != nil {
		return 0, err
	}
	out, err := client.CallContract(ctx, ethereum.CallMsg{To: &addr, Data: data}, nil)
	if err != nil {
		return 0, err
	}

	vals, err := s.poolV3ABI.Unpack("fee", out)
	if err != nil || len(vals) == 0 {
		return 0, fmt.Errorf("unpack fee failed: %w", err)
	}

	switch v := vals[0].(type) {
	case uint32:
		return uint(v), nil
	case uint64:
		return uint(v), nil
	case *big.Int:
		return uint(v.Uint64()), nil
	default:
		return 0, fmt.Errorf("unexpected fee return type")
	}
}

func anyToUint(v any) (uint, error) {
	switch t := v.(type) {
	case uint8:
		return uint(t), nil
	case uint16:
		return uint(t), nil
	case uint32:
		return uint(t), nil
	case uint64:
		return uint(t), nil
	case int8:
		return uint(t), nil
	case int16:
		return uint(t), nil
	case int32:
		return uint(t), nil
	case int64:
		return uint(t), nil
	case *big.Int:
		return uint(t.Uint64()), nil
	default:
		return 0, fmt.Errorf("unexpected uint type")
	}
}

func anyToInt64(v any) (int64, error) {
	switch t := v.(type) {
	case int8:
		return int64(t), nil
	case int16:
		return int64(t), nil
	case int32:
		return int64(t), nil
	case int64:
		return t, nil
	case uint8:
		return int64(t), nil
	case uint16:
		return int64(t), nil
	case uint32:
		return int64(t), nil
	case uint64:
		return int64(t), nil
	case *big.Int:
		return t.Int64(), nil
	default:
		return 0, fmt.Errorf("unexpected int type")
	}
}

func anyToBigInt(v any) (*big.Int, error) {
	switch t := v.(type) {
	case *big.Int:
		return t, nil
	case uint64:
		return new(big.Int).SetUint64(t), nil
	case int64:
		return big.NewInt(t), nil
	default:
		return nil, fmt.Errorf("unexpected big.Int type")
	}
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
