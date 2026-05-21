package repository

import (
	"github.com/alexkalak/whatever-system/src/modules/assets/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AssetMappingRepository interface {
	Create(assetMapping *entities.AssetMapping) error
	GetAll() ([]entities.AssetMapping, error)
	GetByID(id uuid.UUID) (*entities.AssetMapping, error)
	Update(assetMapping *entities.AssetMapping) error
	Delete(id uuid.UUID) error
}

type assetMappingRepository struct {
	db *gorm.DB
}

func NewAssetMappingRepository(db *gorm.DB) AssetMappingRepository {
	return &assetMappingRepository{db: db}
}

func (r *assetMappingRepository) Create(assetMapping *entities.AssetMapping) error {
	return r.db.Create(assetMapping).Error
}

func (r *assetMappingRepository) GetAll() ([]entities.AssetMapping, error) {
	var assetMappings []entities.AssetMapping
	err := r.db.Find(&assetMappings).Error
	return assetMappings, err
}

func (r *assetMappingRepository) GetByID(id uuid.UUID) (*entities.AssetMapping, error) {
	var assetMapping entities.AssetMapping
	err := r.db.First(&assetMapping, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &assetMapping, nil
}

func (r *assetMappingRepository) Update(assetMapping *entities.AssetMapping) error {
	return r.db.Save(assetMapping).Error
}

func (r *assetMappingRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&entities.AssetMapping{}, "id = ?", id).Error
}
