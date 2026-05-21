package service

import (
	"context"
	"errors"
	"strings"

	chaineventbus "github.com/alexkalak/whatever-system/src/modules/eventbus/chain"
	"github.com/alexkalak/whatever-system/src/modules/exchanges/dex/entities"
	"github.com/alexkalak/whatever-system/src/modules/exchanges/dex/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DexChainEventsService interface {
	EnsureDexFromSwapV3(ctx context.Context, payload chaineventbus.SwapV3Payload) error
}

type dexChainEventsService struct {
	repo repository.DexRepository
}

func NewDexChainEventsService(repo repository.DexRepository) DexChainEventsService {
	return &dexChainEventsService{repo: repo}
}

func (s *dexChainEventsService) EnsureDexFromSwapV3(_ context.Context, payload chaineventbus.SwapV3Payload) error {
	address := strings.ToLower(strings.TrimSpace(payload.PoolAddress))
	if address == "" {
		return nil
	}

	_, err := s.repo.GetByChainIDAndAddress(uint64(payload.ChainID), address)
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	dexType := mapDexType(payload.Dex)
	if dexType == "" {
		return nil
	}

	return s.repo.Create(&entities.Dex{
		ID:      uuid.New(),
		DexType: dexType,
		ChainID: uint64(payload.ChainID),
		Address: address,
		FeeTier: 0,
	})
}

func mapDexType(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "uniswap_v3", "uniswapv3":
		return "uniswapv3"
	case "pancakeswap_v3", "pancakeswapv3":
		return "pancakeswapv3"
	case "sushiswap_v3", "sushiswapv3":
		return "sushiswapv3"
	case "uniswap_v2", "uniswapv2":
		return "uniswapv2"
	default:
		return ""
	}
}
