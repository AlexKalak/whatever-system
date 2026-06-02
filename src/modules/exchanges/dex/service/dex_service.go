package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/alexkalak/whatever-system/src/modules/exchanges/dex/entities"
	"github.com/alexkalak/whatever-system/src/modules/exchanges/dex/repository"
	"github.com/google/uuid"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

var dexEnsureGroup singleflight.Group

type DexService interface {
	Create(dex *entities.Dex) (*entities.Dex, error)
	EnsureByChainIDAndAddress(chainID uint64, address string, dexType string) (*entities.Dex, error)
	GetAll() ([]entities.Dex, error)
	GetPaginated(page, limit int) ([]entities.Dex, int64, error)
	GetPaginatedByType(dexType string, page, limit int) ([]entities.Dex, int64, error)
	GetByID(id uuid.UUID) (*entities.Dex, error)
	GetByChainIDAndAddress(chainID uint64, address string) (*entities.Dex, error)
	GetByTypeChainIDAndAddress(dexType string, chainID uint64, address string) (*entities.Dex, error)
	Update(id uuid.UUID, payload *entities.Dex) (*entities.Dex, error)
	Delete(id uuid.UUID) error
}

type dexService struct {
	repo repository.DexRepository
}

func NewDexService(repo repository.DexRepository) DexService {
	return &dexService{repo: repo}
}

func (s *dexService) Create(dex *entities.Dex) (*entities.Dex, error) {
	dex.ID = uuid.New()
	if err := s.repo.Create(dex); err != nil {
		return nil, err
	}
	return dex, nil
}

func (s *dexService) EnsureByChainIDAndAddress(chainID uint64, address string, dexType string) (*entities.Dex, error) {
	address = strings.ToLower(strings.TrimSpace(address))
	dexType = strings.ToLower(strings.TrimSpace(dexType))
	key := fmt.Sprintf("%d:%s", chainID, address)

	result, err, _ := dexEnsureGroup.Do(key, func() (any, error) {
		dex, err := s.repo.GetByChainIDAndAddress(chainID, address)
		if err == nil {
			return dex, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		newDex := &entities.Dex{DexType: dexType, ChainID: chainID, Address: address}
		return s.Create(newDex)
	})
	if err != nil {
		return nil, err
	}
	return result.(*entities.Dex), nil
}

func (s *dexService) GetAll() ([]entities.Dex, error) {
	return s.repo.GetAll()
}

func (s *dexService) GetPaginated(page, limit int) ([]entities.Dex, int64, error) {
	return s.repo.GetPaginated(page, limit)
}

func (s *dexService) GetPaginatedByType(dexType string, page, limit int) ([]entities.Dex, int64, error) {
	return s.repo.GetPaginatedByType(strings.ToLower(strings.TrimSpace(dexType)), page, limit)
}

func (s *dexService) GetByID(id uuid.UUID) (*entities.Dex, error) {
	return s.repo.GetByID(id)
}

func (s *dexService) GetByChainIDAndAddress(chainID uint64, address string) (*entities.Dex, error) {
	return s.repo.GetByChainIDAndAddress(chainID, strings.ToLower(strings.TrimSpace(address)))
}

func (s *dexService) GetByTypeChainIDAndAddress(dexType string, chainID uint64, address string) (*entities.Dex, error) {
	return s.repo.GetByTypeChainIDAndAddress(strings.ToLower(strings.TrimSpace(dexType)), chainID, strings.ToLower(strings.TrimSpace(address)))
}

func (s *dexService) Update(id uuid.UUID, payload *entities.Dex) (*entities.Dex, error) {
	dex, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	dex.DexType = payload.DexType
	dex.ChainID = payload.ChainID
	dex.Address = payload.Address

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
