package service

import (
	"github.com/alexkalak/whatever-system/src/modules/assets/entities"
	"github.com/alexkalak/whatever-system/src/modules/assets/repository"
	"github.com/google/uuid"
)

type AssetMappingService interface {
	Create(assetMapping *entities.AssetMapping) error
	GetAll() ([]entities.AssetMapping, error)
	GetByID(id uuid.UUID) (*entities.AssetMapping, error)
	Update(id uuid.UUID, payload *entities.AssetMapping) (*entities.AssetMapping, error)
	Delete(id uuid.UUID) error
}

type assetMappingService struct {
	repo repository.AssetMappingRepository
}

func NewAssetMappingService(repo repository.AssetMappingRepository) AssetMappingService {
	return &assetMappingService{repo: repo}
}

func (s *assetMappingService) Create(assetMapping *entities.AssetMapping) error {
	assetMapping.ID = uuid.New()
	return s.repo.Create(assetMapping)
}

func (s *assetMappingService) GetAll() ([]entities.AssetMapping, error) {
	return s.repo.GetAll()
}

func (s *assetMappingService) GetByID(id uuid.UUID) (*entities.AssetMapping, error) {
	return s.repo.GetByID(id)
}

func (s *assetMappingService) Update(id uuid.UUID, payload *entities.AssetMapping) (*entities.AssetMapping, error) {
	assetMapping, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	assetMapping.Type = payload.Type
	assetMapping.Name = payload.Name
	assetMapping.Exchange = payload.Exchange
	assetMapping.SymbolOnExchange = payload.SymbolOnExchange
	assetMapping.ChainID = payload.ChainID
	assetMapping.Address = payload.Address
	assetMapping.Decimals = payload.Decimals

	if err := s.repo.Update(assetMapping); err != nil {
		return nil, err
	}

	return assetMapping, nil
}

func (s *assetMappingService) Delete(id uuid.UUID) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	return s.repo.Delete(id)
}
