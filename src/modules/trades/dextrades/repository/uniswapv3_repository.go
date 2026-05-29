package repository

import (
	"github.com/alexkalak/whatever-system/src/modules/trades/dextrades/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DexTradeUniswapV3Repository interface {
	Create(trade *entities.DexTradeUniswapV3) error
	GetAll() ([]entities.DexTradeUniswapV3, error)
	GetByID(id uuid.UUID) (*entities.DexTradeUniswapV3, error)
	GetByChainIDPoolTxHash(chainID uint64, poolAddress, txHash string) (*entities.DexTradeUniswapV3, error)
	Update(trade *entities.DexTradeUniswapV3) error
	Delete(id uuid.UUID) error
}

type dexTradeUniswapV3Repository struct {
	db *gorm.DB
}

func NewDexTradeUniswapV3Repository(db *gorm.DB) DexTradeUniswapV3Repository {
	return &dexTradeUniswapV3Repository{db: db}
}

func (r *dexTradeUniswapV3Repository) Create(trade *entities.DexTradeUniswapV3) error {
	return r.db.Create(trade).Error
}

func (r *dexTradeUniswapV3Repository) GetAll() ([]entities.DexTradeUniswapV3, error) {
	var trades []entities.DexTradeUniswapV3
	err := r.db.Find(&trades).Error
	return trades, err
}

func (r *dexTradeUniswapV3Repository) GetByID(id uuid.UUID) (*entities.DexTradeUniswapV3, error) {
	var trade entities.DexTradeUniswapV3
	err := r.db.First(&trade, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &trade, nil
}

func (r *dexTradeUniswapV3Repository) GetByChainIDPoolTxHash(chainID uint64, poolAddress, txHash string) (*entities.DexTradeUniswapV3, error) {
	var trade entities.DexTradeUniswapV3
	err := r.db.First(&trade, "chain_id = ? AND pool_address = ? AND tx_hash = ?", chainID, poolAddress, txHash).Error
	if err != nil {
		return nil, err
	}
	return &trade, nil
}

func (r *dexTradeUniswapV3Repository) Update(trade *entities.DexTradeUniswapV3) error {
	return r.db.Save(trade).Error
}

func (r *dexTradeUniswapV3Repository) Delete(id uuid.UUID) error {
	return r.db.Delete(&entities.DexTradeUniswapV3{}, "id = ?", id).Error
}
