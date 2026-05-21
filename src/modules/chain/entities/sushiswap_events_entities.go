package entities

import (
	"fmt"
	"math/big"

	"github.com/alexkalak/whatever-system/src/modules/chain/chainerrors"
	"github.com/alexkalak/whatever-system/src/shared/tools/chaintools"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type SushiswapV3SwapEvent struct {
	Sender       common.Address `json:"sender"`
	Recipient    common.Address `json:"recipient"`
	Amount0      *big.Int       `json:"amount0"`
	Amount1      *big.Int       `json:"amount1"`
	SqrtPriceX96 *big.Int       `json:"sqrt_price_x96"`
	Liquidity    *big.Int       `json:"liquidity"`
	Tick         *big.Int       `json:"tick"`
}

func (e *SushiswapV3SwapEvent) FromLog(lg types.Log) error {
	if len(lg.Topics) < 3 {
		return chainerrors.ErrUnableToParseLog
	}

	abi, err := GetSushiswapV3EventsABI()
	if err != nil {
		return err
	}

	out, err := abi.ABI.Unpack("Swap", lg.Data)
	if err != nil {
		fmt.Println("Error unpacking Swap event", err)
		return err
	}

	Amount0, ok := out[0].(*big.Int)
	if !ok {
		return chainerrors.ErrUnableToParseLog
	}
	Amount1, ok := out[1].(*big.Int)
	if !ok {
		return chainerrors.ErrUnableToParseLog
	}
	SqrtPriceX96, ok := out[2].(*big.Int)
	if !ok {
		return chainerrors.ErrUnableToParseLog
	}
	Liquidity, ok := out[3].(*big.Int)
	if !ok {
		return chainerrors.ErrUnableToParseLog
	}
	Tick, ok := out[4].(*big.Int)
	if !ok {
		return chainerrors.ErrUnableToParseLog
	}
	Sender := common.HexToAddress(lg.Topics[1].Hex())
	Recipient := common.HexToAddress(lg.Topics[2].Hex())

	e.Sender = Sender
	e.Recipient = Recipient
	e.Amount0 = Amount0
	e.Amount1 = Amount1
	e.SqrtPriceX96 = SqrtPriceX96
	e.Liquidity = Liquidity
	e.Tick = Tick

	return nil
}

type SushiswapV3MintEvent struct {
	Sender    common.Address `json:"sender"`
	Owner     common.Address `json:"owner"`
	TickLower int32          `json:"tick_lower"`
	TickUpper int32          `json:"tick_upper"`
	Amount    *big.Int       `json:"amount"`
	Amount0   *big.Int       `json:"amount0"`
	Amount1   *big.Int       `json:"amount1"`
}

func (e *SushiswapV3MintEvent) FromLog(lg types.Log) error {
	if len(lg.Topics) < 4 {
		return chainerrors.ErrUnableToParseLog
	}

	abi, err := GetSushiswapV3EventsABI()
	if err != nil {
		return err
	}

	out, err := abi.ABI.Unpack("Mint", lg.Data)
	if err != nil {
		fmt.Println("Error unpacking Mint event", err)
		return err
	}

	Sender, ok := out[0].(common.Address) // int256
	if !ok {
		return chainerrors.ErrUnableToParseLog
	}
	Amount, ok := out[1].(*big.Int)
	if !ok {
		return chainerrors.ErrUnableToParseLog
	}
	Amount0, ok := out[2].(*big.Int)
	if !ok {
		return chainerrors.ErrUnableToParseLog
	}
	Amount1, ok := out[3].(*big.Int)
	if !ok {
		return chainerrors.ErrUnableToParseLog
	}

	Owner := common.HexToAddress(lg.Topics[1].Hex())
	TickLower := chaintools.ParseUniswapV3TickFromTopic(lg.Topics[2])
	TickUpper := chaintools.ParseUniswapV3TickFromTopic(lg.Topics[3])

	e.Sender = Sender
	e.Owner = Owner
	e.TickLower = TickLower
	e.TickUpper = TickUpper
	e.Amount = Amount
	e.Amount0 = Amount0
	e.Amount1 = Amount1

	return nil
}

type SushiswapV3BurnEvent struct {
	Owner     common.Address `json:"owner"`
	TickLower int32          `json:"tick_lower"`
	TickUpper int32          `json:"tick_upper"`
	Amount    *big.Int       `json:"amount"`
	Amount0   *big.Int       `json:"amount0"`
	Amount1   *big.Int       `json:"amount1"`
}

func (e *SushiswapV3BurnEvent) FromLog(lg types.Log) error {
	if len(lg.Topics) < 4 {
		return chainerrors.ErrUnableToParseLog
	}

	abi, err := GetSushiswapV3EventsABI()
	if err != nil {
		return err
	}

	out, err := abi.ABI.Unpack("Burn", lg.Data)
	if err != nil {
		fmt.Println("Error unpacking Burn event", err)
		return err
	}

	Amount, ok := out[0].(*big.Int)
	if !ok {
		return chainerrors.ErrUnableToParseLog
	}
	Amount0, ok := out[1].(*big.Int)
	if !ok {
		return chainerrors.ErrUnableToParseLog
	}
	Amount1, ok := out[2].(*big.Int)
	if !ok {
		return chainerrors.ErrUnableToParseLog
	}

	Owner := common.HexToAddress(lg.Topics[1].Hex())
	TickLower := chaintools.ParseUniswapV3TickFromTopic(lg.Topics[2])
	TickUpper := chaintools.ParseUniswapV3TickFromTopic(lg.Topics[3])

	e.Owner = Owner
	e.TickLower = TickLower
	e.TickUpper = TickUpper
	e.Amount = Amount
	e.Amount0 = Amount0
	e.Amount1 = Amount1

	return nil
}

func (SushiswapV3SwapEvent) isChainEventData() {}
func (SushiswapV3MintEvent) isChainEventData() {}
func (SushiswapV3BurnEvent) isChainEventData() {}
