package service

import (
	"slices"

	"github.com/alexkalak/whatever-system/src/modules/chain/chainerrors"
	chainentities "github.com/alexkalak/whatever-system/src/modules/chain/entities"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type chainLogsParser struct {
	uniswapV3ABI      *chainentities.V3ExchangeABI
	pancakeswapV3ABI  *chainentities.V3ExchangeABI
	sushiswapV3ABI    *chainentities.V3ExchangeABI
	uniswapV2ABI      *chainentities.V2ExchangeABI
	uniswapV3Sigs     []common.Hash
	pancakeswapV3Sigs []common.Hash
	sushiswapV3Sigs   []common.Hash
	uniswapV2Sigs     []common.Hash
}

func newChainLogsParser() (chainLogsParser, error) {
	uniswapV2ABI, err := chainentities.GetUniswapV2EventsABI()
	if err != nil {
		return chainLogsParser{}, err
	}
	uniswapV3ABI, err := chainentities.GetUniswapV3EventsABI()
	if err != nil {
		return chainLogsParser{}, err
	}
	pancakeswapV3ABI, err := chainentities.GetPancakeswapV3EventsABI()
	if err != nil {
		return chainLogsParser{}, err
	}
	sushiswapV3ABI, err := chainentities.GetSushiswapV3EventsABI()
	if err != nil {
		return chainLogsParser{}, err
	}

	return chainLogsParser{uniswapV2ABI: uniswapV2ABI, uniswapV3ABI: uniswapV3ABI, pancakeswapV3ABI: pancakeswapV3ABI, sushiswapV3ABI: sushiswapV3ABI, uniswapV2Sigs: uniswapV2ABI.GetSigs(), uniswapV3Sigs: uniswapV3ABI.GetSigs(), pancakeswapV3Sigs: pancakeswapV3ABI.GetSigs(), sushiswapV3Sigs: sushiswapV3ABI.GetSigs()}, err
}

func (c *chainLogsParser) parse(lg types.Log) (chainentities.ChainEventType, chainentities.ChainEventData, error) {
	if len(lg.Topics) == 0 {
		return "", nil, chainerrors.ErrEventSigNotFound
	}
	if slices.Contains(c.uniswapV2Sigs, lg.Topics[0]) {
		switch lg.Topics[0] {
		case c.uniswapV2ABI.SwapV2Sig:
			e := chainentities.UniswapV2SwapEvent{}
			if err := e.FromLog(lg); err != nil {
				return "", nil, err
			}
			return chainentities.ChainSwapV2Event, e, nil
		case c.uniswapV2ABI.SyncV2Sig:
			e := chainentities.UniswapV2SyncEvent{}
			if err := e.FromLog(lg); err != nil {
				return "", nil, err
			}
			return chainentities.ChainSyncV2Event, e, nil
		}
	}
	if slices.Contains(c.uniswapV3Sigs, lg.Topics[0]) {
		switch lg.Topics[0] {
		case c.uniswapV3ABI.SwapV3Sig:
			e := chainentities.UniswapV3SwapEvent{}
			if err := e.FromLog(lg); err != nil {
				return "", nil, err
			}
			return chainentities.ChainSwapV3Event, e, nil
		case c.uniswapV3ABI.MintV3Sig:
			e := chainentities.UniswapV3MintEvent{}
			if err := e.FromLog(lg); err != nil {
				return "", nil, err
			}
			return chainentities.ChainMintV3Event, e, nil
		case c.uniswapV3ABI.BurnV3Sig:
			e := chainentities.UniswapV3BurnEvent{}
			if err := e.FromLog(lg); err != nil {
				return "", nil, err
			}
			return chainentities.ChainBurnV3Event, e, nil
		}
	}
	if slices.Contains(c.pancakeswapV3Sigs, lg.Topics[0]) {
		switch lg.Topics[0] {
		case c.pancakeswapV3ABI.SwapV3Sig:
			e := chainentities.PancakeswapV3SwapEvent{}
			if err := e.FromLog(lg); err != nil {
				return "", nil, err
			}
			return chainentities.ChainSwapV3Event, e, nil
		case c.pancakeswapV3ABI.MintV3Sig:
			e := chainentities.PancakeswapV3MintEvent{}
			if err := e.FromLog(lg); err != nil {
				return "", nil, err
			}
			return chainentities.ChainMintV3Event, e, nil
		case c.pancakeswapV3ABI.BurnV3Sig:
			e := chainentities.PancakeswapV3BurnEvent{}
			if err := e.FromLog(lg); err != nil {
				return "", nil, err
			}
			return chainentities.ChainBurnV3Event, e, nil
		}
	}
	if slices.Contains(c.sushiswapV3Sigs, lg.Topics[0]) {
		switch lg.Topics[0] {
		case c.sushiswapV3ABI.SwapV3Sig:
			e := chainentities.SushiswapV3SwapEvent{}
			if err := e.FromLog(lg); err != nil {
				return "", nil, err
			}
			return chainentities.ChainSwapV3Event, e, nil
		case c.sushiswapV3ABI.MintV3Sig:
			e := chainentities.SushiswapV3MintEvent{}
			if err := e.FromLog(lg); err != nil {
				return "", nil, err
			}
			return chainentities.ChainMintV3Event, e, nil
		case c.sushiswapV3ABI.BurnV3Sig:
			e := chainentities.SushiswapV3BurnEvent{}
			if err := e.FromLog(lg); err != nil {
				return "", nil, err
			}
			return chainentities.ChainBurnV3Event, e, nil
		}
	}
	return "", nil, chainerrors.ErrEventSigNotFound
}
