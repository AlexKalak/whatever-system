package repository

import (
	"github.com/alexkalak/whatever-system/src/modules/exchanges/dex/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UniswapV3Repository interface {
	Create(item *entities.UniswapV3Dex) error
	Update(item *entities.UniswapV3Dex) error
	GetByDexID(dexID uuid.UUID) (*entities.UniswapV3Dex, error)
	GetIncomplete() ([]entities.UniswapV3Dex, error)
}

type uniswapV3Repository struct {
	db *gorm.DB
}

func NewUniswapV3Repository(db *gorm.DB) UniswapV3Repository {
	return &uniswapV3Repository{db: db}
}

func (r *uniswapV3Repository) Create(item *entities.UniswapV3Dex) error {
	return r.db.Create(item).Error
}

func (r *uniswapV3Repository) Update(item *entities.UniswapV3Dex) error {
	return r.db.Save(item).Error
}

func (r *uniswapV3Repository) GetByDexID(dexID uuid.UUID) (*entities.UniswapV3Dex, error) {
	var item entities.UniswapV3Dex
	err := r.db.First(&item, "dex_id = ?", dexID).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *uniswapV3Repository) GetIncomplete() ([]entities.UniswapV3Dex, error) {
	var items []entities.UniswapV3Dex
	err := r.db.
		Preload("Dex").
		Where("token0_address = '' OR token1_address = '' OR fee_tier = 0 OR tick_spacing = 0").
		Find(&items).Error
	return items, err
}
