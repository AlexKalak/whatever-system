package entities

import (
	"time"

	"github.com/alexkalak/whatever-system/src/shared/tools/dbtypes"
	"github.com/google/uuid"
)

type DexTradeUniswapV2 struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	ChainID     uint64         `json:"chainId" gorm:"type:bigint;not null;index"`
	DexAddress  string         `json:"dexAddress" gorm:"type:varchar(255);not null;index"`
	BlockNumber uint64         `json:"blockNumber" gorm:"type:bigint;not null;index"`
	PoolAddress string         `json:"poolAddress" gorm:"type:varchar(255);not null;index"`
	TxHash      string         `json:"txHash" gorm:"type:varchar(255);not null;index"`
	Sender      string         `json:"sender" gorm:"type:varchar(255);not null;index"`
	Recipient   string         `json:"recipient" gorm:"type:varchar(255);not null;index"`
	Amount0In   dbtypes.BigInt `json:"amount0In" gorm:"type:numeric;not null"`
	Amount1In   dbtypes.BigInt `json:"amount1In" gorm:"type:numeric;not null"`
	Amount0Out  dbtypes.BigInt `json:"amount0Out" gorm:"type:numeric;not null"`
	Amount1Out  dbtypes.BigInt `json:"amount1Out" gorm:"type:numeric;not null"`
	TradePrice  dbtypes.BigInt `json:"tradePrice" gorm:"type:numeric;not null"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type DexTradeUniswapV3 struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	ChainID      uint64         `json:"chainId" gorm:"type:bigint;not null;index"`
	DexAddress   string         `json:"dexAddress" gorm:"type:varchar(255);not null;index"`
	BlockNumber  uint64         `json:"blockNumber" gorm:"type:bigint;not null;index"`
	PoolAddress  string         `json:"poolAddress" gorm:"type:varchar(255);not null;index"`
	TxHash       string         `json:"txHash" gorm:"type:varchar(255);not null;index"`
	Sender       string         `json:"sender" gorm:"type:varchar(255);not null;index"`
	Recipient    string         `json:"recipient" gorm:"type:varchar(255);not null;index"`
	Amount0      dbtypes.BigInt `json:"amount0" gorm:"type:numeric;not null"`
	Amount1      dbtypes.BigInt `json:"amount1" gorm:"type:numeric;not null"`
	SqrtPriceX96 dbtypes.BigInt `json:"sqrtPriceX96" gorm:"type:numeric;not null"`
	Liquidity    dbtypes.BigInt `json:"liquidity" gorm:"type:numeric;not null"`
	Tick         int64          `json:"tick" gorm:"type:bigint;not null"`
	TradePrice   dbtypes.BigInt `json:"tradePrice" gorm:"type:numeric;not null"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
