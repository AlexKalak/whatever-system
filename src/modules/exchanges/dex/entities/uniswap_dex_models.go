package entities

import (
	"github.com/alexkalak/whatever-system/src/shared/tools/dbtypes"
	"github.com/google/uuid"
)

type UniswapV2Dex struct {
	DexID uuid.UUID `json:"dexId" gorm:"type:uuid;primaryKey"`
	Dex   Dex       `json:"dex" gorm:"foreignKey:DexID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	Token0Address string `json:"token0Address" gorm:"type:varchar(255);not null;index"`
	Token1Address string `json:"token1Address" gorm:"type:varchar(255);not null;index"`
	Token0Amount  string `json:"token0Amount" gorm:"type:numeric;not null"`
	Token1Amount  string `json:"token1Amount" gorm:"type:numeric;not null"`
	FeeTier       uint   `json:"feeTier" gorm:"type:uint"`
	ExchangeName  string `json:"exchangeName" gorm:"type:varchar(30);not null;index"` // uniswap | pancakeswap | sushiswap
}

type UniswapV3Dex struct {
	DexID uuid.UUID `json:"dexId" gorm:"type:uuid;primaryKey"`
	Dex   Dex       `json:"dex" gorm:"foreignKey:DexID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	Token0Address string         `json:"token0Address" gorm:"type:varchar(255);not null;index"`
	Token1Address string         `json:"token1Address" gorm:"type:varchar(255);not null;index"`
	FeeTier       uint           `json:"feeTier" gorm:"type:uint"`
	SqrtPriceX96  dbtypes.BigInt `json:"sqrtPriceX96" gorm:"type:numeric;not null"`
	Liquidity     dbtypes.BigInt `json:"liquidity" gorm:"type:numeric;not null"`
	Tick          int64          `json:"tick" gorm:"type:bigint;not null"`
	TickSpacing   int64          `json:"tickSpacing" gorm:"type:bigint;not null"`
}
