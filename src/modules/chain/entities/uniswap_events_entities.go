package entities

import (
	"fmt"
	"math/big"

	"github.com/alexkalak/whatever-system/src/modules/chain/chainerrors"
	"github.com/alexkalak/whatever-system/src/shared/tools/chaintools"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type UniswapV3SwapEvent struct {
	Sender       common.Address `json:"sender"`
	Recipient    common.Address `json:"recipient"`
	Amount0      *big.Int       `json:"amount0"`
	Amount1      *big.Int       `json:"amount1"`
	SqrtPriceX96 *big.Int       `json:"sqrt_price_x96"`
	Liquidity    *big.Int       `json:"liquidity"`
	Tick         *big.Int       `json:"tick"`
}

func (e *UniswapV3SwapEvent) FromLog(lg types.Log) error {
	if len(lg.Topics) < 3 {
		return chainerrors.ErrUnableToParseLog
	}

	abi, err := GetUniswapV3EventsABI()
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

type UniswapV3MintEvent struct {
	Sender    common.Address `json:"sender"`
	Owner     common.Address `json:"owner"`
	TickLower int32          `json:"tick_lower"`
	TickUpper int32          `json:"tick_upper"`
	Amount    *big.Int       `json:"amount"`
	Amount0   *big.Int       `json:"amount0"`
	Amount1   *big.Int       `json:"amount1"`
}

func (e *UniswapV3MintEvent) FromLog(lg types.Log) error {
	if len(lg.Topics) < 4 {
		return chainerrors.ErrUnableToParseLog
	}

	abi, err := GetUniswapV3EventsABI()
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

type UniswapV3BurnEvent struct {
	Owner     common.Address `json:"owner"`
	TickLower int32          `json:"tick_lower"`
	TickUpper int32          `json:"tick_upper"`
	Amount    *big.Int       `json:"amount"`
	Amount0   *big.Int       `json:"amount0"`
	Amount1   *big.Int       `json:"amount1"`
}

func (e *UniswapV3BurnEvent) FromLog(lg types.Log) error {
	if len(lg.Topics) < 4 {
		return chainerrors.ErrUnableToParseLog
	}

	abi, err := GetUniswapV3EventsABI()
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

type UniswapV2SwapEvent struct {
	Sender     common.Address `json:"sender"`
	Amount0In  *big.Int       `json:"amount0_in"`
	Amount1In  *big.Int       `json:"amount1_in"`
	Amount0Out *big.Int       `json:"amount0_out"`
	Amount1Out *big.Int       `json:"amount1_out"`
	To         common.Address `json:"to"`
}

func (e *UniswapV2SwapEvent) FromLog(lg types.Log) error {
	if len(lg.Topics) < 3 {
		return chainerrors.ErrUnableToParseLog
	}

	abi, err := GetUniswapV2EventsABI()
	if err != nil {
		return err
	}

	out, err := abi.ABI.Unpack("Swap", lg.Data)
	if err != nil {
		fmt.Println("Error unpacking V2 Swap event", err)
		return err
	}

	amount0In, ok := out[0].(*big.Int)
	if !ok {
		return chainerrors.ErrUnableToParseLog
	}
	amount1In, ok := out[1].(*big.Int)
	if !ok {
		return chainerrors.ErrUnableToParseLog
	}
	amount0Out, ok := out[2].(*big.Int)
	if !ok {
		return chainerrors.ErrUnableToParseLog
	}
	amount1Out, ok := out[3].(*big.Int)
	if !ok {
		return chainerrors.ErrUnableToParseLog
	}

	sender := common.HexToAddress(lg.Topics[1].Hex())
	to := common.HexToAddress(lg.Topics[2].Hex())

	e.Sender = sender
	e.To = to
	e.Amount0In = amount0In
	e.Amount1In = amount1In
	e.Amount0Out = amount0Out
	e.Amount1Out = amount1Out

	return nil
}

type UniswapV2SyncEvent struct {
	Reserve0 *big.Int `json:"reserve0"`
	Reserve1 *big.Int `json:"reserve1"`
}

func (e *UniswapV2SyncEvent) FromLog(lg types.Log) error {
	abi, err := GetUniswapV2EventsABI()
	if err != nil {
		return err
	}

	out, err := abi.ABI.Unpack("Sync", lg.Data)
	if err != nil {
		fmt.Println("Error unpacking V2 Sync event", err)
		return err
	}

	reserve0, ok := out[0].(*big.Int)
	if !ok {
		return chainerrors.ErrUnableToParseLog
	}
	reserve1, ok := out[1].(*big.Int)
	if !ok {
		return chainerrors.ErrUnableToParseLog
	}

	e.Reserve0 = reserve0
	e.Reserve1 = reserve1

	return nil
}

func (UniswapV3SwapEvent) isChainEventData() {}
func (UniswapV3MintEvent) isChainEventData() {}
func (UniswapV3BurnEvent) isChainEventData() {}
func (UniswapV2SwapEvent) isChainEventData() {}
func (UniswapV2SyncEvent) isChainEventData() {}
