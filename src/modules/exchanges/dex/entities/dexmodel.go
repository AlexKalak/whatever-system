package entities

import (
	"time"

	"github.com/google/uuid"
)

type Dex struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`

	DexType string `json:"dexType" gorm:"type:varchar(30);not null;index"` // uniswapv2 | uniswapv3 | pancakeswapv3 | sushiswapv3
	ChainID uint64 `json:"chainId" gorm:"type:bigint;not null;index"`
	Address string `json:"address" gorm:"type:varchar(255);not null;uniqueIndex"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
