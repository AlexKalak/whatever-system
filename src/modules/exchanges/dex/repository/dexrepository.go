package repository

import (
	"github.com/alexkalak/whatever-system/src/modules/exchanges/dex/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DexRepository interface {
	Create(dex *entities.Dex) error
	GetAll() ([]entities.Dex, error)
	GetByID(id uuid.UUID) (*entities.Dex, error)
	GetByChainIDAndAddress(chainID uint64, address string) (*entities.Dex, error)
	Update(dex *entities.Dex) error
	Delete(id uuid.UUID) error
}

type dexRepository struct {
	db *gorm.DB
}

func NewDexRepository(db *gorm.DB) DexRepository {
	return &dexRepository{db: db}
}

func (r *dexRepository) Create(dex *entities.Dex) error {
	return r.db.Create(dex).Error
}

func (r *dexRepository) GetAll() ([]entities.Dex, error) {
	var dexes []entities.Dex
	err := r.db.Find(&dexes).Error
	return dexes, err
}

func (r *dexRepository) GetByID(id uuid.UUID) (*entities.Dex, error) {
	var dex entities.Dex
	err := r.db.First(&dex, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &dex, nil
}

func (r *dexRepository) GetByChainIDAndAddress(chainID uint64, address string) (*entities.Dex, error) {
	var dex entities.Dex
	err := r.db.First(&dex, "chain_id = ? AND address = ?", chainID, address).Error
	if err != nil {
		return nil, err
	}
	return &dex, nil
}

func (r *dexRepository) Update(dex *entities.Dex) error {
	return r.db.Save(dex).Error
}

func (r *dexRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&entities.Dex{}, "id = ?", id).Error
}
