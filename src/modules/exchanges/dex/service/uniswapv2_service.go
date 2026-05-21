package service

import (
	"context"
	"errors"
	"strings"

	dexentities "github.com/alexkalak/whatever-system/src/modules/exchanges/dex/entities"
	dexrepository "github.com/alexkalak/whatever-system/src/modules/exchanges/dex/repository"
	"github.com/alexkalak/whatever-system/src/modules/tokens/service"
	"gorm.io/gorm"
)

type UniswapV2DexService interface {
	Create(item *dexentities.UniswapV2Dex) error
}

type uniswapV2DexService struct {
	repo         dexrepository.UniswapV2Repository
	dexRepo      dexrepository.DexRepository
	tokenService service.TokenService
}

func NewUniswapV2DexService(repo dexrepository.UniswapV2Repository, dexRepo dexrepository.DexRepository, tokenService service.TokenService) UniswapV2DexService {
	return &uniswapV2DexService{repo: repo, dexRepo: dexRepo, tokenService: tokenService}
}

func (s *uniswapV2DexService) Create(item *dexentities.UniswapV2Dex) error {
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

	if err := s.tokenService.EnsureTokenExists(context.Background(), uint(dex.ChainID), item.Token0Address); err != nil {
		return err
	}
	if err := s.tokenService.EnsureTokenExists(context.Background(), uint(dex.ChainID), item.Token1Address); err != nil {
		return err
	}

	item.Token0Address = strings.ToLower(strings.TrimSpace(item.Token0Address))
	item.Token1Address = strings.ToLower(strings.TrimSpace(item.Token1Address))
	item.ExchangeName = strings.ToLower(strings.TrimSpace(item.ExchangeName))

	return s.repo.Create(item)
}
