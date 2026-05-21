package repository

import (
	"github.com/alexkalak/whatever-system/src/modules/assets/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AssetRepository interface {
	Create(asset *entities.Asset) error
	GetAll() ([]entities.Asset, error)
	GetByID(id uuid.UUID) (*entities.Asset, error)
	Update(asset *entities.Asset) error
	Delete(id uuid.UUID) error
}

type assetRepository struct {
	db *gorm.DB
}

func NewAssetRepository(db *gorm.DB) AssetRepository {
	return &assetRepository{db: db}
}

func (r *assetRepository) Create(asset *entities.Asset) error {
	return r.db.Create(asset).Error
}

func (r *assetRepository) GetAll() ([]entities.Asset, error) {
	var assets []entities.Asset
	err := r.db.Find(&assets).Error
	return assets, err
}

func (r *assetRepository) GetByID(id uuid.UUID) (*entities.Asset, error) {
	var asset entities.Asset
	err := r.db.First(&asset, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *assetRepository) Update(asset *entities.Asset) error {
	return r.db.Save(asset).Error
}

func (r *assetRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&entities.Asset{}, "id = ?", id).Error
}
