package entities

import (
	"strings"

	"github.com/alexkalak/whatever-system/src/shared/chaindata/abis"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type V3ExchangeABI struct {
	ABI        abi.ABI
	SwapV3Sig  common.Hash
	MintV3Sig  common.Hash
	BurnV3Sig  common.Hash
	FlashV3Sig common.Hash
}

func (v *V3ExchangeABI) GetSigs() []common.Hash {
	return []common.Hash{
		v.SwapV3Sig,
		v.MintV3Sig,
		v.BurnV3Sig,
		v.FlashV3Sig,
	}
}

var uniswapV3EventsABI *V3ExchangeABI
var sushiswapV3EventsABI *V3ExchangeABI
var pancakeswapV3EventsABI *V3ExchangeABI

func GetUniswapV3EventsABI() (*V3ExchangeABI, error) {
	if uniswapV3EventsABI != nil {
		return uniswapV3EventsABI, nil
	}

	abi, err := abi.JSON(strings.NewReader(abis.EventsABIUniswapV3String))
	if err != nil {
		return nil, err
	}

	uniswapV3EventsABI = &V3ExchangeABI{
		ABI:        abi,
		SwapV3Sig:  abi.Events["Swap"].ID,
		MintV3Sig:  abi.Events["Mint"].ID,
		BurnV3Sig:  abi.Events["Burn"].ID,
		FlashV3Sig: abi.Events["Flash"].ID,
	}

	return uniswapV3EventsABI, nil
}

func GetSushiswapV3EventsABI() (*V3ExchangeABI, error) {
	if sushiswapV3EventsABI != nil {
		return sushiswapV3EventsABI, nil
	}

	abi, err := abi.JSON(strings.NewReader(abis.EventsABISushiswapV3String))
	if err != nil {
		return nil, err
	}

	sushiswapV3EventsABI = &V3ExchangeABI{
		ABI:        abi,
		SwapV3Sig:  abi.Events["Swap"].ID,
		MintV3Sig:  abi.Events["Mint"].ID,
		BurnV3Sig:  abi.Events["Burn"].ID,
		FlashV3Sig: abi.Events["Flash"].ID,
	}

	return sushiswapV3EventsABI, nil
}

func GetPancakeswapV3EventsABI() (*V3ExchangeABI, error) {
	if pancakeswapV3EventsABI != nil {
		return pancakeswapV3EventsABI, nil
	}

	abi, err := abi.JSON(strings.NewReader(abis.EventsABIPancakeswapV3String))
	if err != nil {
		return nil, err
	}

	pancakeswapV3EventsABI = &V3ExchangeABI{
		ABI:        abi,
		SwapV3Sig:  abi.Events["Swap"].ID,
		MintV3Sig:  abi.Events["Mint"].ID,
		BurnV3Sig:  abi.Events["Burn"].ID,
		FlashV3Sig: abi.Events["Flash"].ID,
	}

	return pancakeswapV3EventsABI, nil
}

type V2ExchangeABI struct {
	ABI       abi.ABI
	SyncV2Sig common.Hash
	SwapV2Sig common.Hash
}

func (v *V2ExchangeABI) GetSigs() []common.Hash {
	return []common.Hash{
		v.SwapV2Sig,
		v.SyncV2Sig,
	}
}

var uniswapV2EventsABI *V2ExchangeABI

func GetUniswapV2EventsABI() (*V2ExchangeABI, error) {
	if uniswapV2EventsABI != nil {
		return uniswapV2EventsABI, nil
	}

	abi, err := abi.JSON(strings.NewReader(abis.EventsABIUniswapV2String))
	if err != nil {
		return nil, err
	}

	uniswapV2EventsABI = &V2ExchangeABI{
		ABI:       abi,
		SyncV2Sig: abi.Events["Sync"].ID,
		SwapV2Sig: abi.Events["Swap"].ID,
	}

	return uniswapV2EventsABI, nil
}
