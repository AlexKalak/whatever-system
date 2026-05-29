package repository

import (
	"github.com/alexkalak/whatever-system/src/modules/exchanges/dex/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UniswapV2Repository interface {
	Create(item *entities.UniswapV2Dex) error
	Update(item *entities.UniswapV2Dex) error
	GetByDexID(dexID uuid.UUID) (*entities.UniswapV2Dex, error)
	GetIncomplete() ([]entities.UniswapV2Dex, error)
}

type uniswapV2Repository struct {
	db *gorm.DB
}

func NewUniswapV2Repository(db *gorm.DB) UniswapV2Repository {
	return &uniswapV2Repository{db: db}
}

func (r *uniswapV2Repository) Create(item *entities.UniswapV2Dex) error {
	return r.db.Create(item).Error
}

func (r *uniswapV2Repository) Update(item *entities.UniswapV2Dex) error {
	return r.db.Save(item).Error
}

func (r *uniswapV2Repository) GetByDexID(dexID uuid.UUID) (*entities.UniswapV2Dex, error) {
	var item entities.UniswapV2Dex
	err := r.db.First(&item, "dex_id = ?", dexID).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *uniswapV2Repository) GetIncomplete() ([]entities.UniswapV2Dex, error) {
	var items []entities.UniswapV2Dex
	err := r.db.
		Preload("Dex").
		Where("token0_address = '' OR token1_address = ''").
		Find(&items).Error
	return items, err
}
