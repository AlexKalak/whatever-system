package repository

import (
	"github.com/alexkalak/whatever-system/src/modules/trades/dextrades/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DexTradeUniswapV2Repository interface {
	Create(trade *entities.DexTradeUniswapV2) error
	GetAll() ([]entities.DexTradeUniswapV2, error)
	GetByID(id uuid.UUID) (*entities.DexTradeUniswapV2, error)
	GetByChainIDPoolTxHash(chainID uint64, poolAddress, txHash string) (*entities.DexTradeUniswapV2, error)
	Update(trade *entities.DexTradeUniswapV2) error
	Delete(id uuid.UUID) error
}

type dexTradeUniswapV2Repository struct {
	db *gorm.DB
}

func NewDexTradeUniswapV2Repository(db *gorm.DB) DexTradeUniswapV2Repository {
	return &dexTradeUniswapV2Repository{db: db}
}

func (r *dexTradeUniswapV2Repository) Create(trade *entities.DexTradeUniswapV2) error {
	return r.db.Create(trade).Error
}

func (r *dexTradeUniswapV2Repository) GetAll() ([]entities.DexTradeUniswapV2, error) {
	var trades []entities.DexTradeUniswapV2
	err := r.db.Find(&trades).Error
	return trades, err
}

func (r *dexTradeUniswapV2Repository) GetByID(id uuid.UUID) (*entities.DexTradeUniswapV2, error) {
	var trade entities.DexTradeUniswapV2
	err := r.db.First(&trade, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &trade, nil
}

func (r *dexTradeUniswapV2Repository) GetByChainIDPoolTxHash(chainID uint64, poolAddress, txHash string) (*entities.DexTradeUniswapV2, error) {
	var trade entities.DexTradeUniswapV2
	err := r.db.First(&trade, "chain_id = ? AND pool_address = ? AND tx_hash = ?", chainID, poolAddress, txHash).Error
	if err != nil {
		return nil, err
	}
	return &trade, nil
}

func (r *dexTradeUniswapV2Repository) Update(trade *entities.DexTradeUniswapV2) error {
	return r.db.Save(trade).Error
}

func (r *dexTradeUniswapV2Repository) Delete(id uuid.UUID) error {
	return r.db.Delete(&entities.DexTradeUniswapV2{}, "id = ?", id).Error
}
