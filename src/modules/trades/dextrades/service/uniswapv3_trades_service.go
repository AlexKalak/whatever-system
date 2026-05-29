package service

import (
	"context"
	"fmt"
	"strings"

	dexrepository "github.com/alexkalak/whatever-system/src/modules/exchanges/dex/repository"
	tokenservice "github.com/alexkalak/whatever-system/src/modules/tokens/service"
	"github.com/alexkalak/whatever-system/src/modules/trades/dextrades/entities"
	"github.com/alexkalak/whatever-system/src/modules/trades/dextrades/repository"
	"github.com/alexkalak/whatever-system/src/shared/tools/chaintools"
	"github.com/google/uuid"
)

const tradePriceOutputDecimals uint8 = 18

type DexTradeUniswapV3Service interface {
	Create(trade *entities.DexTradeUniswapV3) (*entities.DexTradeUniswapV3, error)
	GetAll() ([]entities.DexTradeUniswapV3, error)
	GetByID(id uuid.UUID) (*entities.DexTradeUniswapV3, error)
	Update(id uuid.UUID, payload *entities.DexTradeUniswapV3) (*entities.DexTradeUniswapV3, error)
	UpdateTradePrice(ctx context.Context, id uuid.UUID) (*entities.DexTradeUniswapV3, error)
	Delete(id uuid.UUID) error
}

type dexTradeUniswapV3Service struct {
	repo          repository.DexTradeUniswapV3Repository
	dexRepo       dexrepository.DexRepository
	uniswapV3Repo dexrepository.UniswapV3Repository
	tokenService  tokenservice.TokenService
	priceDecimals uint8
}

func NewDexTradeUniswapV3Service(repo repository.DexTradeUniswapV3Repository, deps ...DexTradeUniswapV3ServiceDeps) DexTradeUniswapV3Service {
	service := &dexTradeUniswapV3Service{repo: repo, priceDecimals: tradePriceOutputDecimals}
	if len(deps) > 0 {
		service.dexRepo = deps[0].DexRepo
		service.uniswapV3Repo = deps[0].UniswapV3Repo
		service.tokenService = deps[0].TokenService
		if deps[0].PriceDecimals != 0 {
			service.priceDecimals = deps[0].PriceDecimals
		}
	}
	return service
}

// DexTradeUniswapV3ServiceDeps are optional dependencies used by UpdateTradePrice.
type DexTradeUniswapV3ServiceDeps struct {
	DexRepo       dexrepository.DexRepository
	UniswapV3Repo dexrepository.UniswapV3Repository
	TokenService  tokenservice.TokenService
	PriceDecimals uint8
}

func (s *dexTradeUniswapV3Service) Create(trade *entities.DexTradeUniswapV3) (*entities.DexTradeUniswapV3, error) {
	trade.PoolAddress = strings.ToLower(strings.TrimSpace(trade.PoolAddress))
	trade.TxHash = strings.ToLower(strings.TrimSpace(trade.TxHash))
	if existing, err := s.repo.GetByChainIDPoolTxHash(trade.ChainID, trade.PoolAddress, trade.TxHash); err == nil {
		return existing, nil
	} else if !tokenservice.IsNotFound(err) {
		return nil, err
	}

	trade.ID = uuid.New()
	if err := s.repo.Create(trade); err != nil {
		return nil, err
	}
	return trade, nil
}

func (s *dexTradeUniswapV3Service) GetAll() ([]entities.DexTradeUniswapV3, error) {
	return s.repo.GetAll()
}

func (s *dexTradeUniswapV3Service) GetByID(id uuid.UUID) (*entities.DexTradeUniswapV3, error) {
	return s.repo.GetByID(id)
}

func (s *dexTradeUniswapV3Service) Update(id uuid.UUID, payload *entities.DexTradeUniswapV3) (*entities.DexTradeUniswapV3, error) {
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
	trade.Amount0 = payload.Amount0
	trade.Amount1 = payload.Amount1
	trade.SqrtPriceX96 = payload.SqrtPriceX96
	trade.Liquidity = payload.Liquidity
	trade.Tick = payload.Tick
	trade.TradePrice = payload.TradePrice

	if err := s.repo.Update(trade); err != nil {
		return nil, err
	}
	return trade, nil
}

func (s *dexTradeUniswapV3Service) UpdateTradePrice(ctx context.Context, id uuid.UUID) (*entities.DexTradeUniswapV3, error) {
	if s.dexRepo == nil || s.uniswapV3Repo == nil || s.tokenService == nil {
		return nil, fmt.Errorf("update trade price dependencies are not configured")
	}

	trade, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	poolAddress := strings.ToLower(strings.TrimSpace(trade.PoolAddress))
	if poolAddress == "" {
		return nil, fmt.Errorf("trade %s has empty pool address", id)
	}

	dex, err := s.dexRepo.GetByChainIDAndAddress(trade.ChainID, poolAddress)
	if err != nil {
		return nil, err
	}

	pool, err := s.uniswapV3Repo.GetByDexID(dex.ID)
	if err != nil {
		return nil, err
	}

	token0Address := strings.ToLower(strings.TrimSpace(pool.Token0Address))
	token1Address := strings.ToLower(strings.TrimSpace(pool.Token1Address))
	if token0Address == "" || token1Address == "" {
		return nil, fmt.Errorf("pool %s tokens are not loaded", poolAddress)
	}

	chainID := uint(trade.ChainID)
	if err := s.tokenService.EnsureTokenExists(ctx, chainID, token0Address); err != nil {
		return nil, err
	}
	if err := s.tokenService.EnsureTokenExists(ctx, chainID, token1Address); err != nil {
		return nil, err
	}

	token0, err := s.tokenService.GetByChainIDAndAddress(chainID, token0Address)
	if err != nil {
		return nil, err
	}
	token1, err := s.tokenService.GetByChainIDAndAddress(chainID, token1Address)
	if err != nil {
		return nil, err
	}

	tradePrice, err := chaintools.CountTradePriceBigIntBySqrtPriceX96AndDecimals(trade.SqrtPriceX96, token0.Decimals, token1.Decimals, s.priceDecimals)
	if err != nil {
		return nil, err
	}

	trade.TradePrice = tradePrice
	if err := s.repo.Update(trade); err != nil {
		return nil, err
	}

	return trade, nil
}

func (s *dexTradeUniswapV3Service) Delete(id uuid.UUID) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	return s.repo.Delete(id)
}
