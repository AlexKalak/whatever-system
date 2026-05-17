package service

import (
	"errors"

	"github.com/alexkalak/whatever-system/src/modules/assets/entities"
	"github.com/alexkalak/whatever-system/src/modules/assets/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AssetService interface {
	Create(asset *entities.Asset) error
	GetAll() ([]entities.Asset, error)
	GetByID(id uuid.UUID) (*entities.Asset, error)
	Update(id uuid.UUID, payload *entities.Asset) (*entities.Asset, error)
	Delete(id uuid.UUID) error
}

type assetService struct {
	repo repository.AssetRepository
}

func NewAssetService(repo repository.AssetRepository) AssetService {
	return &assetService{repo: repo}
}

func (s *assetService) Create(asset *entities.Asset) error {
	asset.ID = uuid.New()
	return s.repo.Create(asset)
}

func (s *assetService) GetAll() ([]entities.Asset, error) {
	return s.repo.GetAll()
}

func (s *assetService) GetByID(id uuid.UUID) (*entities.Asset, error) {
	return s.repo.GetByID(id)
}

func (s *assetService) Update(id uuid.UUID, payload *entities.Asset) (*entities.Asset, error) {
	asset, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	asset.Name = payload.Name
	asset.Symbol = payload.Symbol
	asset.Type = payload.Type

	if err := s.repo.Update(asset); err != nil {
		return nil, err
	}

	return asset, nil
}

func (s *assetService) Delete(id uuid.UUID) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	return s.repo.Delete(id)
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
