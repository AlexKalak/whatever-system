package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	chainservice "github.com/alexkalak/whatever-system/src/modules/chain/service"
	dexentities "github.com/alexkalak/whatever-system/src/modules/exchanges/dex/entities"
	dexrepository "github.com/alexkalak/whatever-system/src/modules/exchanges/dex/repository"
	"github.com/alexkalak/whatever-system/src/modules/tokens/service"
	"github.com/alexkalak/whatever-system/src/shared/tools/dbtypes"
	"github.com/google/uuid"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

var uniswapV3EnsureGroup singleflight.Group

type UniswapV3DexService interface {
	Create(item *dexentities.UniswapV3Dex) error
	Update(item *dexentities.UniswapV3Dex) error
	EnsureByDex(ctx context.Context, dex *dexentities.Dex) (*dexentities.UniswapV3Dex, error)
	ensureExists(item *dexentities.UniswapV3Dex) (*dexentities.UniswapV3Dex, error)
	GetByDexID(dexID uuid.UUID) (*dexentities.UniswapV3Dex, error)
}

type uniswapV3DexService struct {
	repo         dexrepository.UniswapV3Repository
	dexRepo      dexrepository.DexRepository
	tokenService service.TokenService
	chainData    chainservice.ChainDataService
}

func NewUniswapV3DexService(repo dexrepository.UniswapV3Repository, dexRepo dexrepository.DexRepository, tokenService service.TokenService, chainData chainservice.ChainDataService) UniswapV3DexService {
	return &uniswapV3DexService{repo: repo, dexRepo: dexRepo, tokenService: tokenService, chainData: chainData}
}

func (s *uniswapV3DexService) Create(item *dexentities.UniswapV3Dex) error {
	if err := s.prepare(item); err != nil {
		return err
	}

	return s.repo.Create(item)
}

func (s *uniswapV3DexService) Update(item *dexentities.UniswapV3Dex) error {
	return s.repo.Update(item)
}

func (s *uniswapV3DexService) GetByDexID(dexID uuid.UUID) (*dexentities.UniswapV3Dex, error) {
	return s.repo.GetByDexID(dexID)
}

func (s *uniswapV3DexService) EnsureByDex(ctx context.Context, dex *dexentities.Dex) (*dexentities.UniswapV3Dex, error) {
	key := fmt.Sprintf("%d:%s", dex.ChainID, strings.ToLower(strings.TrimSpace(dex.Address)))
	result, err, _ := uniswapV3EnsureGroup.Do(key, func() (any, error) {
		if existing, err := s.GetByDexID(dex.ID); err == nil {
			return existing, nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		log.Println("Getting v3PoolInfo while ensuring: ", dex.ChainID, dex.Address)
		info, err := s.chainData.GetUniswapV3PoolInfo(ctx, uint(dex.ChainID), dex.Address)
		log.Println("Got v3PoolInfo for ensuring: ", dex.ChainID, dex.Address, info.Token0Address, info.Token1Address)
		if err != nil {
			return nil, err
		}
		sqrtPriceX96, err := dbtypes.NewBigIntFromString(info.SqrtPriceX96)
		if err != nil {
			return nil, err
		}
		liquidity, err := dbtypes.NewBigIntFromString(info.Liquidity)
		if err != nil {
			return nil, err
		}
		return s.ensureExists(&dexentities.UniswapV3Dex{
			DexID:         dex.ID,
			Dex:           *dex,
			Token0Address: info.Token0Address,
			Token1Address: info.Token1Address,
			FeeTier:       info.FeeTier,
			SqrtPriceX96:  sqrtPriceX96,
			Liquidity:     liquidity,
			Tick:          info.Tick,
			TickSpacing:   info.TickSpacing,
		})
	})
	if err != nil {
		return nil, err
	}
	return result.(*dexentities.UniswapV3Dex), nil
}

func (s *uniswapV3DexService) ensureExists(item *dexentities.UniswapV3Dex) (*dexentities.UniswapV3Dex, error) {
	existing, err := s.GetByDexID(item.DexID)
	if err == nil {
		if err := s.prepare(item); err != nil {
			return nil, err
		}

		existing.Token0Address = item.Token0Address
		existing.Token1Address = item.Token1Address
		existing.FeeTier = item.FeeTier
		existing.TickSpacing = item.TickSpacing

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

func (s *uniswapV3DexService) prepare(item *dexentities.UniswapV3Dex) error {
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
