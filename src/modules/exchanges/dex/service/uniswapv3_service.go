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

type UniswapV3DexService interface {
	Create(item *dexentities.UniswapV3Dex) error
}

type uniswapV3DexService struct {
	repo         dexrepository.UniswapV3Repository
	dexRepo      dexrepository.DexRepository
	tokenService service.TokenService
}

func NewUniswapV3DexService(repo dexrepository.UniswapV3Repository, dexRepo dexrepository.DexRepository, tokenService service.TokenService) UniswapV3DexService {
	return &uniswapV3DexService{repo: repo, dexRepo: dexRepo, tokenService: tokenService}
}

func (s *uniswapV3DexService) Create(item *dexentities.UniswapV3Dex) error {
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

	if err := s.tokenService.EnsureTokenExists(context.Background(), uint(dex.ChainID), item.Token0Address); err != nil {
		return err
	}
	if err := s.tokenService.EnsureTokenExists(context.Background(), uint(dex.ChainID), item.Token1Address); err != nil {
		return err
	}

	item.Token0Address = strings.ToLower(strings.TrimSpace(item.Token0Address))
	item.Token1Address = strings.ToLower(strings.TrimSpace(item.Token1Address))

	return s.repo.Create(item)
}

