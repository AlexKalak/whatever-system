package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	chainservice "github.com/alexkalak/whatever-system/src/modules/chain/service"
	dexentities "github.com/alexkalak/whatever-system/src/modules/exchanges/dex/entities"
	dexrepository "github.com/alexkalak/whatever-system/src/modules/exchanges/dex/repository"
	"github.com/alexkalak/whatever-system/src/modules/tokens/service"
	"github.com/google/uuid"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

var uniswapV2EnsureGroup singleflight.Group

type UniswapV2DexService interface {
	Create(item *dexentities.UniswapV2Dex) error
	Update(item *dexentities.UniswapV2Dex) error
	EnsureByDex(ctx context.Context, dex *dexentities.Dex) (*dexentities.UniswapV2Dex, error)
	ensureExists(item *dexentities.UniswapV2Dex) (*dexentities.UniswapV2Dex, error)
	GetByDexID(dexID uuid.UUID) (*dexentities.UniswapV2Dex, error)
}

type uniswapV2DexService struct {
	repo         dexrepository.UniswapV2Repository
	dexRepo      dexrepository.DexRepository
	tokenService service.TokenService
	chainData    chainservice.ChainDataService
}

func NewUniswapV2DexService(repo dexrepository.UniswapV2Repository, dexRepo dexrepository.DexRepository, tokenService service.TokenService, chainData chainservice.ChainDataService) UniswapV2DexService {
	return &uniswapV2DexService{repo: repo, dexRepo: dexRepo, tokenService: tokenService, chainData: chainData}
}

func (s *uniswapV2DexService) Create(item *dexentities.UniswapV2Dex) error {
	if err := s.prepare(item); err != nil {
		return err
	}

	return s.repo.Create(item)
}

func (s *uniswapV2DexService) Update(item *dexentities.UniswapV2Dex) error {
	return s.repo.Update(item)
}

func (s *uniswapV2DexService) GetByDexID(dexID uuid.UUID) (*dexentities.UniswapV2Dex, error) {
	return s.repo.GetByDexID(dexID)
}

func (s *uniswapV2DexService) EnsureByDex(ctx context.Context, dex *dexentities.Dex) (*dexentities.UniswapV2Dex, error) {
	key := fmt.Sprintf("%d:%s", dex.ChainID, strings.ToLower(strings.TrimSpace(dex.Address)))
	result, err, _ := uniswapV2EnsureGroup.Do(key, func() (any, error) {
		if existing, err := s.GetByDexID(dex.ID); err == nil {
			return existing, nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		info, err := s.chainData.GetUniswapV2PairInfo(ctx, uint(dex.ChainID), dex.Address)
		if err != nil {
			return nil, err
		}
		return s.ensureExists(&dexentities.UniswapV2Dex{
			DexID:         dex.ID,
			Dex:           *dex,
			Token0Address: info.Token0Address,
			Token1Address: info.Token1Address,
			Token0Amount:  info.Reserve0,
			Token1Amount:  info.Reserve1,
			FeeTier:       info.FeeTier,
			ExchangeName:  "uniswap",
		})
	})
	if err != nil {
		return nil, err
	}
	return result.(*dexentities.UniswapV2Dex), nil
}

func (s *uniswapV2DexService) ensureExists(item *dexentities.UniswapV2Dex) (*dexentities.UniswapV2Dex, error) {
	existing, err := s.GetByDexID(item.DexID)
	if err == nil {
		if err := s.prepare(item); err != nil {
			return nil, err
		}

		existing.Token0Address = item.Token0Address
		existing.Token1Address = item.Token1Address
		existing.FeeTier = item.FeeTier
		existing.ExchangeName = item.ExchangeName

		if err := s.repo.Update(existing); err != nil {
			return nil, err
		}
		return existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if err := s.Create(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *uniswapV2DexService) prepare(item *dexentities.UniswapV2Dex) error {
	dex, err := s.dexRepo.GetByID(item.DexID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			item.Dex.ID = item.DexID
			if err := s.dexRepo.Create(&item.Dex); err != nil {
				return err
			}
			dex = &item.Dex
		} else {
			return err
		}
	}

	item.Dex = *dex

	item.Token0Address = strings.ToLower(strings.TrimSpace(item.Token0Address))
	item.Token1Address = strings.ToLower(strings.TrimSpace(item.Token1Address))
	item.ExchangeName = strings.ToLower(strings.TrimSpace(item.ExchangeName))

	if item.Token0Address != "" {
		if err := s.tokenService.EnsureTokenExists(context.Background(), uint(dex.ChainID), item.Token0Address); err != nil {
			return err
		}
	}
	if item.Token1Address != "" {
		if err := s.tokenService.EnsureTokenExists(context.Background(), uint(dex.ChainID), item.Token1Address); err != nil {
			return err
		}
	}

	return nil
}
