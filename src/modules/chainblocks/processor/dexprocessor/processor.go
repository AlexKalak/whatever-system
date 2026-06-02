package dexprocessor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	chainentities "github.com/alexkalak/whatever-system/src/modules/chain/entities"
	chainpubsub "github.com/alexkalak/whatever-system/src/modules/chain/pubsub"
	chainservice "github.com/alexkalak/whatever-system/src/modules/chain/service"
	dexentities "github.com/alexkalak/whatever-system/src/modules/exchanges/dex/entities"
	dexrepository "github.com/alexkalak/whatever-system/src/modules/exchanges/dex/repository"
	dexservice "github.com/alexkalak/whatever-system/src/modules/exchanges/dex/service"
	tokenrepository "github.com/alexkalak/whatever-system/src/modules/tokens/repository"
	tokenservice "github.com/alexkalak/whatever-system/src/modules/tokens/service"
	"github.com/alexkalak/whatever-system/src/shared/tools/dbtypes"
	"gorm.io/gorm"
)

type Processor struct {
	invalidDexes map[string]any
	chainData    chainservice.ChainDataService
	dexService   dexservice.DexService
	v2DexService dexservice.UniswapV2DexService
	v3DexService dexservice.UniswapV3DexService
}

type Deps struct {
	ChainData     chainservice.ChainDataService
	DexRepo       dexrepository.DexRepository
	UniswapV2Repo dexrepository.UniswapV2Repository
	UniswapV3Repo dexrepository.UniswapV3Repository
	TokenService  tokenservice.TokenService
}

func New(db *gorm.DB, deps Deps) *Processor {
	tokenSvc := deps.TokenService
	if tokenSvc == nil {
		tokenSvc = tokenservice.NewTokenService(tokenrepository.NewTokenRepository(db), deps.ChainData)
	}

	return &Processor{
		invalidDexes: make(map[string]any),
		chainData:    deps.ChainData,
		dexService:   dexservice.NewDexService(deps.DexRepo),
		v2DexService: dexservice.NewUniswapV2DexService(deps.UniswapV2Repo, deps.DexRepo, tokenSvc, deps.ChainData),
		v3DexService: dexservice.NewUniswapV3DexService(deps.UniswapV3Repo, deps.DexRepo, tokenSvc, deps.ChainData),
	}
}

func (p *Processor) ProcessEvent(ctx context.Context, block chainpubsub.ChainBlockMessage) error {
	for _, event := range block.Events {
		err := p.processEvent(ctx, block.ChainID, event)
		if err == nil {
			continue
		}
		log.Println("Err processing dex: ", event.Address, err, block.ChainID)
		p.invalidDexes[event.Address] = new(any)
	}
	return nil
}

func (p *Processor) processEvent(ctx context.Context, chainID uint, event chainpubsub.ChainEventMessage) error {
	switch event.Type {
	case chainentities.ChainSwapV2Event, chainentities.ChainSyncV2Event:
		return p.processV2DexEvent(ctx, chainID, event)
	case chainentities.ChainSwapV3Event, chainentities.ChainMintV3Event, chainentities.ChainBurnV3Event:
		return p.processV3DexEvent(ctx, chainID, event)
	default:
		return nil
	}
}

func (p *Processor) processV2DexEvent(ctx context.Context, chainID uint, event chainpubsub.ChainEventMessage) error {
	v2Dex, err := p.ensureV2Dex(ctx, chainID, event.Address)
	if err != nil {
		return err
	}
	if event.Type != chainentities.ChainSyncV2Event {
		return nil
	}

	var data chainentities.UniswapV2SyncEvent
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return err
	}

	v2Dex.Token0Amount = data.Reserve0.String()
	v2Dex.Token1Amount = data.Reserve1.String()
	return p.v2DexService.Update(v2Dex)
}

func (p *Processor) processV3DexEvent(ctx context.Context, chainID uint, event chainpubsub.ChainEventMessage) error {
	v3Dex, err := p.ensureV3Dex(ctx, chainID, event.Address)
	if err != nil {
		return err
	}
	switch event.Type {
	case chainentities.ChainSwapV3Event:
		var data chainentities.UniswapV3SwapEvent
		if err := json.Unmarshal(event.Data, &data); err != nil {
			return err
		}

		v3Dex.SqrtPriceX96 = dbtypes.NewBigInt(data.SqrtPriceX96)
		v3Dex.Liquidity = dbtypes.NewBigInt(data.Liquidity)
		v3Dex.Tick = data.Tick.Int64()
		return p.v3DexService.Update(v3Dex)
	case chainentities.ChainMintV3Event, chainentities.ChainBurnV3Event:
		fmt.Println("Handling mint/burn event, ", event.Address)
		return p.refreshV3DexFromChain(ctx, chainID, event.Address, v3Dex)
	default:
		return nil
	}
}

func (p *Processor) refreshV3DexFromChain(ctx context.Context, chainID uint, address string, v3Dex *dexentities.UniswapV3Dex) error {
	info, err := p.chainData.GetUniswapV3PoolInfo(ctx, chainID, address)
	if err != nil {
		return err
	}
	sqrtPriceX96, err := dbtypes.NewBigIntFromString(info.SqrtPriceX96)
	if err != nil {
		return err
	}
	liquidity, err := dbtypes.NewBigIntFromString(info.Liquidity)
	if err != nil {
		return err
	}

	v3Dex.Token0Address = info.Token0Address
	v3Dex.Token1Address = info.Token1Address
	v3Dex.FeeTier = info.FeeTier
	v3Dex.SqrtPriceX96 = sqrtPriceX96
	v3Dex.Liquidity = liquidity
	v3Dex.Tick = info.Tick
	v3Dex.TickSpacing = info.TickSpacing
	return p.v3DexService.Update(v3Dex)
}

func (p *Processor) ensureV2Dex(ctx context.Context, chainID uint, address string) (*dexentities.UniswapV2Dex, error) {
	address = strings.ToLower(strings.TrimSpace(address))
	dex, err := p.dexService.EnsureByChainIDAndAddress(uint64(chainID), address, "uniswapv2")
	if err != nil {
		return nil, err
	}
	return p.v2DexService.EnsureByDex(ctx, dex)
}

func (p *Processor) ensureV3Dex(ctx context.Context, chainID uint, address string) (*dexentities.UniswapV3Dex, error) {
	address = strings.ToLower(strings.TrimSpace(address))
	dex, err := p.dexService.EnsureByChainIDAndAddress(uint64(chainID), address, "uniswapv3")
	if err != nil {
		return nil, err
	}
	return p.v3DexService.EnsureByDex(ctx, dex)
}
