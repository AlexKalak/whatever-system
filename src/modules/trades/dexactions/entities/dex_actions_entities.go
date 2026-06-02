package entities

import (
	"time"

	"github.com/alexkalak/whatever-system/src/shared/tools/dbtypes"
	"github.com/google/uuid"
)

const (
	ActionTypeSwap = "swap"
	ActionTypeMint = "mint"
	ActionTypeBurn = "burn"
)

type DexActionUniswapV2 struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	ChainID       uint64         `json:"chainId" gorm:"type:bigint;not null;index"`
	DexAddress    string         `json:"dexAddress" gorm:"type:varchar(255);not null;index"`
	ActionType    string         `json:"actionType" gorm:"type:varchar(30);not null;index"`
	BlockNumber   uint64         `json:"blockNumber" gorm:"type:bigint;not null;index"`
	IndexInBlock  uint64         `json:"indexInBlock" gorm:"type:bigint;not null;default:0;index"`
	IndexInTx     uint64         `json:"indexInTx" gorm:"type:bigint;not null;default:0;index"`
	PoolAddress   string         `json:"poolAddress" gorm:"type:varchar(255);not null;index"`
	TxHash        string         `json:"txHash" gorm:"type:varchar(255);not null;index"`
	Amount0       dbtypes.BigInt `json:"amount0" gorm:"type:numeric;not null"`
	Amount1       dbtypes.BigInt `json:"amount1" gorm:"type:numeric;not null"`
	Metadata      dbtypes.JSON   `json:"metadata" gorm:"type:jsonb;not null;default:'{}'"`
	SeenInMempool bool           `json:"seenInMempool" gorm:"column:seen_in_mempool;->;-:migration"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
}

type DexActionUniswapV3 struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	ChainID       uint64         `json:"chainId" gorm:"type:bigint;not null;index"`
	DexAddress    string         `json:"dexAddress" gorm:"type:varchar(255);not null;index"`
	ActionType    string         `json:"actionType" gorm:"type:varchar(30);not null;index"`
	BlockNumber   uint64         `json:"blockNumber" gorm:"type:bigint;not null;index"`
	IndexInBlock  uint64         `json:"indexInBlock" gorm:"type:bigint;not null;default:0;index"`
	IndexInTx     uint64         `json:"indexInTx" gorm:"type:bigint;not null;default:0;index"`
	PoolAddress   string         `json:"poolAddress" gorm:"type:varchar(255);not null;index"`
	TxHash        string         `json:"txHash" gorm:"type:varchar(255);not null;index"`
	Amount0       dbtypes.BigInt `json:"amount0" gorm:"type:numeric;not null"`
	Amount1       dbtypes.BigInt `json:"amount1" gorm:"type:numeric;not null"`
	Metadata      dbtypes.JSON   `json:"metadata" gorm:"type:jsonb;not null;default:'{}'"`
	SeenInMempool bool           `json:"seenInMempool" gorm:"column:seen_in_mempool;->;-:migration"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
}
