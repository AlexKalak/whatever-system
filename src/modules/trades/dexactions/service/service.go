package service

import (
	"errors"
	"strings"

	"github.com/alexkalak/whatever-system/src/modules/trades/dexactions/entities"
	"github.com/alexkalak/whatever-system/src/modules/trades/dexactions/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UniswapV2Service interface {
	Create(action *entities.DexActionUniswapV2) (*entities.DexActionUniswapV2, error)
	GetPaginated(page, limit int, orderBy, direction string) ([]entities.DexActionUniswapV2, int64, error)
	GetByID(id uuid.UUID) (*entities.DexActionUniswapV2, error)
	GetByTxHash(txHash string, page, limit int, orderBy, direction string) ([]entities.DexActionUniswapV2, int64, error)
	GetByChainIDAndDexAddress(chainID uint64, dexAddress string, page, limit int, orderBy, direction string) ([]entities.DexActionUniswapV2, int64, error)
	GetByChainIDAndDexAddressCursor(chainID uint64, dexAddress string, cursor string, limit int, orderBy, direction string) ([]entities.DexActionUniswapV2, int64, string, error)
}

type UniswapV3Service interface {
	Create(action *entities.DexActionUniswapV3) (*entities.DexActionUniswapV3, error)
	GetPaginated(page, limit int, orderBy, direction string) ([]entities.DexActionUniswapV3, int64, error)
	GetByID(id uuid.UUID) (*entities.DexActionUniswapV3, error)
	GetByTxHash(txHash string, page, limit int, orderBy, direction string) ([]entities.DexActionUniswapV3, int64, error)
	GetByChainIDAndDexAddress(chainID uint64, dexAddress string, page, limit int, orderBy, direction string) ([]entities.DexActionUniswapV3, int64, error)
	GetByChainIDAndDexAddressCursor(chainID uint64, dexAddress string, cursor string, limit int, orderBy, direction string) ([]entities.DexActionUniswapV3, int64, string, error)
}

type uniswapV2Service struct {
	repo repository.UniswapV2Repository
}
type uniswapV3Service struct {
	repo repository.UniswapV3Repository
}

func NewUniswapV2Service(repo repository.UniswapV2Repository) UniswapV2Service {
	return &uniswapV2Service{repo: repo}
}
func NewUniswapV3Service(repo repository.UniswapV3Repository) UniswapV3Service {
	return &uniswapV3Service{repo: repo}
}

func (s *uniswapV2Service) Create(action *entities.DexActionUniswapV2) (*entities.DexActionUniswapV2, error) {
	normalizeV2(action)
	if existing, err := s.repo.GetByChainIDPoolTxHashAndIndexes(action.ChainID, action.PoolAddress, action.TxHash, action.ActionType, action.IndexInBlock, action.IndexInTx); err == nil {
		return existing, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	action.ID = uuid.New()
	if err := s.repo.Create(action); err != nil {
		return nil, err
	}
	return action, nil
}

func (s *uniswapV3Service) Create(action *entities.DexActionUniswapV3) (*entities.DexActionUniswapV3, error) {
	normalizeV3(action)
	if existing, err := s.repo.GetByChainIDPoolTxHashActionAndIndexes(action.ChainID, action.PoolAddress, action.TxHash, action.ActionType, action.IndexInBlock, action.IndexInTx); err == nil {
		return existing, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	action.ID = uuid.New()
	if err := s.repo.Create(action); err != nil {
		return nil, err
	}
	return action, nil
}

func (s *uniswapV2Service) GetPaginated(page, limit int, orderBy, direction string) ([]entities.DexActionUniswapV2, int64, error) {
	return s.repo.GetPaginated(page, limit, normalizeOrderBy(orderBy), normalizeDirection(direction))
}
func (s *uniswapV3Service) GetPaginated(page, limit int, orderBy, direction string) ([]entities.DexActionUniswapV3, int64, error) {
	return s.repo.GetPaginated(page, limit, normalizeOrderBy(orderBy), normalizeDirection(direction))
}
func (s *uniswapV2Service) GetByID(id uuid.UUID) (*entities.DexActionUniswapV2, error) {
	return s.repo.GetByID(id)
}
func (s *uniswapV3Service) GetByID(id uuid.UUID) (*entities.DexActionUniswapV3, error) {
	return s.repo.GetByID(id)
}
func (s *uniswapV2Service) GetByTxHash(txHash string, page, limit int, orderBy, direction string) ([]entities.DexActionUniswapV2, int64, error) {
	return s.repo.GetByTxHash(norm(txHash), page, limit, normalizeOrderBy(orderBy), normalizeDirection(direction))
}
func (s *uniswapV3Service) GetByTxHash(txHash string, page, limit int, orderBy, direction string) ([]entities.DexActionUniswapV3, int64, error) {
	return s.repo.GetByTxHash(norm(txHash), page, limit, normalizeOrderBy(orderBy), normalizeDirection(direction))
}
func (s *uniswapV2Service) GetByChainIDAndDexAddress(chainID uint64, dexAddress string, page, limit int, orderBy, direction string) ([]entities.DexActionUniswapV2, int64, error) {
	return s.repo.GetByChainIDAndDexAddress(chainID, norm(dexAddress), page, limit, normalizeOrderBy(orderBy), normalizeDirection(direction))
}
func (s *uniswapV3Service) GetByChainIDAndDexAddress(chainID uint64, dexAddress string, page, limit int, orderBy, direction string) ([]entities.DexActionUniswapV3, int64, error) {
	return s.repo.GetByChainIDAndDexAddress(chainID, norm(dexAddress), page, limit, normalizeOrderBy(orderBy), normalizeDirection(direction))
}
func (s *uniswapV2Service) GetByChainIDAndDexAddressCursor(chainID uint64, dexAddress string, cursor string, limit int, orderBy, direction string) ([]entities.DexActionUniswapV2, int64, string, error) {
	return s.repo.GetByChainIDAndDexAddressCursor(chainID, norm(dexAddress), cursor, limit, normalizeOrderBy(orderBy), normalizeDirection(direction))
}
func (s *uniswapV3Service) GetByChainIDAndDexAddressCursor(chainID uint64, dexAddress string, cursor string, limit int, orderBy, direction string) ([]entities.DexActionUniswapV3, int64, string, error) {
	return s.repo.GetByChainIDAndDexAddressCursor(chainID, norm(dexAddress), cursor, limit, normalizeOrderBy(orderBy), normalizeDirection(direction))
}

func normalizeV2(action *entities.DexActionUniswapV2) {
	action.DexAddress = norm(action.DexAddress)
	action.PoolAddress = norm(action.PoolAddress)
	action.TxHash = norm(action.TxHash)
	action.ActionType = norm(action.ActionType)
}
func normalizeV3(action *entities.DexActionUniswapV3) {
	action.DexAddress = norm(action.DexAddress)
	action.PoolAddress = norm(action.PoolAddress)
	action.TxHash = norm(action.TxHash)
	action.ActionType = norm(action.ActionType)
}
func norm(value string) string { return strings.ToLower(strings.TrimSpace(value)) }

func normalizeOrderBy(orderBy string) string {
	switch norm(orderBy) {
	case "amount0", "amount1":
		return norm(orderBy)
	default:
		return "time"
	}
}
func normalizeDirection(direction string) string {
	if norm(direction) == "asc" {
		return "asc"
	}
	return "desc"
}

func IsNotFound(err error) bool { return errors.Is(err, gorm.ErrRecordNotFound) }
