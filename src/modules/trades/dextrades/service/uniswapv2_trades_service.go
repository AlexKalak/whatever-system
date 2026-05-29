package service

import (
	"errors"
	"strings"

	"github.com/alexkalak/whatever-system/src/modules/trades/dextrades/entities"
	"github.com/alexkalak/whatever-system/src/modules/trades/dextrades/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DexTradeUniswapV2Service interface {
	Create(trade *entities.DexTradeUniswapV2) (*entities.DexTradeUniswapV2, error)
	GetAll() ([]entities.DexTradeUniswapV2, error)
	GetByID(id uuid.UUID) (*entities.DexTradeUniswapV2, error)
	Update(id uuid.UUID, payload *entities.DexTradeUniswapV2) (*entities.DexTradeUniswapV2, error)
	Delete(id uuid.UUID) error
}

type dexTradeUniswapV2Service struct {
	repo repository.DexTradeUniswapV2Repository
}

func NewDexTradeUniswapV2Service(repo repository.DexTradeUniswapV2Repository) DexTradeUniswapV2Service {
	return &dexTradeUniswapV2Service{repo: repo}
}

func (s *dexTradeUniswapV2Service) Create(trade *entities.DexTradeUniswapV2) (*entities.DexTradeUniswapV2, error) {
	trade.PoolAddress = strings.ToLower(strings.TrimSpace(trade.PoolAddress))
	trade.TxHash = strings.ToLower(strings.TrimSpace(trade.TxHash))
	if existing, err := s.repo.GetByChainIDPoolTxHash(trade.ChainID, trade.PoolAddress, trade.TxHash); err == nil {
		return existing, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	trade.ID = uuid.New()
	if err := s.repo.Create(trade); err != nil {
		return nil, err
	}
	return trade, nil
}

func (s *dexTradeUniswapV2Service) GetAll() ([]entities.DexTradeUniswapV2, error) {
	return s.repo.GetAll()
}

func (s *dexTradeUniswapV2Service) GetByID(id uuid.UUID) (*entities.DexTradeUniswapV2, error) {
	return s.repo.GetByID(id)
}

func (s *dexTradeUniswapV2Service) Update(id uuid.UUID, payload *entities.DexTradeUniswapV2) (*entities.DexTradeUniswapV2, error) {
	trade, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	trade.ChainID = payload.ChainID
	trade.DexAddress = payload.DexAddress
	trade.BlockNumber = payload.BlockNumber
	trade.PoolAddress = payload.PoolAddress
	trade.TxHash = payload.TxHash
	trade.Sender = payload.Sender
	trade.Recipient = payload.Recipient
	trade.Amount0In = payload.Amount0In
	trade.Amount1In = payload.Amount1In
	trade.Amount0Out = payload.Amount0Out
	trade.Amount1Out = payload.Amount1Out
	trade.TradePrice = payload.TradePrice

	if err := s.repo.Update(trade); err != nil {
		return nil, err
	}
	return trade, nil
}

func (s *dexTradeUniswapV2Service) Delete(id uuid.UUID) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	return s.repo.Delete(id)
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
