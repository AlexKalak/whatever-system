package service

import (
	"errors"

	"github.com/alexkalak/whatever-system/src/modules/exchanges/dex/entities"
	"github.com/alexkalak/whatever-system/src/modules/exchanges/dex/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DexService interface {
	Create(dex *entities.Dex) error
	GetAll() ([]entities.Dex, error)
	GetByID(id uuid.UUID) (*entities.Dex, error)
	Update(id uuid.UUID, payload *entities.Dex) (*entities.Dex, error)
	Delete(id uuid.UUID) error
}

type dexService struct {
	repo repository.DexRepository
}

func NewDexService(repo repository.DexRepository) DexService {
	return &dexService{repo: repo}
}

func (s *dexService) Create(dex *entities.Dex) error {
	dex.ID = uuid.New()
	return s.repo.Create(dex)
}

func (s *dexService) GetAll() ([]entities.Dex, error) {
	return s.repo.GetAll()
}

func (s *dexService) GetByID(id uuid.UUID) (*entities.Dex, error) {
	return s.repo.GetByID(id)
}

func (s *dexService) Update(id uuid.UUID, payload *entities.Dex) (*entities.Dex, error) {
	dex, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	dex.DexType = payload.DexType
	dex.ChainID = payload.ChainID
	dex.Address = payload.Address
	dex.FeeTier = payload.FeeTier

	if err := s.repo.Update(dex); err != nil {
		return nil, err
	}

	return dex, nil
}

func (s *dexService) Delete(id uuid.UUID) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	return s.repo.Delete(id)
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
