package processor

import (
	"context"
	"encoding/json"
	"errors"
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
	dextradeentities "github.com/alexkalak/whatever-system/src/modules/trades/dextrades/entities"
	dextraderepository "github.com/alexkalak/whatever-system/src/modules/trades/dextrades/repository"
	dextradeservice "github.com/alexkalak/whatever-system/src/modules/trades/dextrades/service"
	"github.com/alexkalak/whatever-system/src/shared/tools/dbtypes"
	"gorm.io/gorm"
)

type CreatedDexTrade struct {
	Version string `json:"version"`
	Trade   any    `json:"trade"`
}

type Processor struct {
	chainData    chainservice.ChainDataService
	dexService   dexservice.DexService
	v2DexService dexservice.UniswapV2DexService
	v3DexService dexservice.UniswapV3DexService
	v2TradeSvc   dextradeservice.DexTradeUniswapV2Service
	v3TradeSvc   dextradeservice.DexTradeUniswapV3Service
}

func New(db *gorm.DB, rpcByChainID map[uint]string) (*Processor, error) {
	chainData, err := chainservice.NewEVMChainDataService(rpcByChainID)
	if err != nil {
		return nil, err
	}

	dexRepo := dexrepository.NewDexRepository(db)
	v2DexRepo := dexrepository.NewUniswapV2Repository(db)
	v3DexRepo := dexrepository.NewUniswapV3Repository(db)
	tokenRepo := tokenrepository.NewTokenRepository(db)
	tokenSvc := tokenservice.NewTokenService(tokenRepo, chainData)

	return &Processor{
		chainData:    chainData,
		dexService:   dexservice.NewDexService(dexRepo),
		v2DexService: dexservice.NewUniswapV2DexService(v2DexRepo, dexRepo, tokenSvc, chainData),
		v3DexService: dexservice.NewUniswapV3DexService(v3DexRepo, dexRepo, tokenSvc, chainData),
		v2TradeSvc:   dextradeservice.NewDexTradeUniswapV2Service(dextraderepository.NewDexTradeUniswapV2Repository(db)),
		v3TradeSvc: dextradeservice.NewDexTradeUniswapV3Service(dextraderepository.NewDexTradeUniswapV3Repository(db), dextradeservice.DexTradeUniswapV3ServiceDeps{
			DexRepo:       dexRepo,
			UniswapV3Repo: v3DexRepo,
			TokenService:  tokenSvc,
		}),
	}, nil
}

func (p *Processor) Process(ctx context.Context, block chainpubsub.ChainBlockMessage) error {
	log.Println("Received block length: ", len(block.Events))
	for _, event := range block.Events {
		for range 3 {
			err := p.processEvent(ctx, block.ChainID, event)
			if err == nil {
				break
			}
			log.Println("Err processing dex: ", err)
		}
	}
	return nil
}

func (p *Processor) EnsureDexesFromBlock(ctx context.Context, block chainpubsub.ChainBlockMessage) error {
	return p.Process(ctx, block)
}

func (p *Processor) CreateTradesFromBlock(ctx context.Context, block chainpubsub.ChainBlockMessage) ([]CreatedDexTrade, error) {
	trades := make([]CreatedDexTrade, 0)
	for _, event := range block.Events {
		for range 3 {
			trade, err := p.createTradeFromEvent(ctx, block.ChainID, event)
			if err == nil {
				if trade != nil {
					trades = append(trades, *trade)
				}
				break
			}
			log.Println("Err creating trades: ", err)
		}
	}
	return trades, nil
}

func (p *Processor) ensureDexFromEvent(ctx context.Context, chainID uint, event chainpubsub.ChainEventMessage) error {
	switch event.Type {
	case chainentities.ChainSwapV2Event, chainentities.ChainSyncV2Event:
		_, err := p.ensureV2Dex(ctx, chainID, event.Address)
		return err
	case chainentities.ChainSwapV3Event, chainentities.ChainMintV3Event, chainentities.ChainBurnV3Event:
		_, err := p.ensureV3Dex(ctx, chainID, event.Address)
		return err
	default:
		return nil
	}
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
	if event.Type != chainentities.ChainSwapV3Event {
		return nil
	}

	var data chainentities.UniswapV3SwapEvent
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return err
	}

	v3Dex.SqrtPriceX96 = dbtypes.NewBigInt(data.SqrtPriceX96)
	v3Dex.Liquidity = dbtypes.NewBigInt(data.Liquidity)
	v3Dex.Tick = data.Tick.Int64()
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

func (p *Processor) createTradeFromEvent(ctx context.Context, chainID uint, event chainpubsub.ChainEventMessage) (*CreatedDexTrade, error) {
	log.Println("Creating trade from event: ", chainID, event.Address)
	switch event.Type {
	case chainentities.ChainSwapV2Event:
		return p.createV2Trade(ctx, chainID, event)
	case chainentities.ChainSwapV3Event:
		return p.createV3Trade(ctx, chainID, event)
	default:
		return nil, errors.New("Unknown chain event")
	}
}

func (p *Processor) createV2Trade(ctx context.Context, chainID uint, event chainpubsub.ChainEventMessage) (*CreatedDexTrade, error) {
	if _, err := p.ensureV2Dex(ctx, chainID, event.Address); err != nil {
		return nil, err
	}
	var data chainentities.UniswapV2SwapEvent
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return nil, err
	}
	log.Println("Creating V2: ", chainID, event.Address, data.Sender.Hex(), data.To.Hex())
	created, err := p.v2TradeSvc.Create(&dextradeentities.DexTradeUniswapV2{
		ChainID:     uint64(chainID),
		DexAddress:  strings.ToLower(event.Address),
		BlockNumber: event.BlockNumber,
		PoolAddress: strings.ToLower(event.Address),
		TxHash:      strings.ToLower(event.TxHash),
		Sender:      strings.ToLower(data.Sender.Hex()),
		Recipient:   strings.ToLower(data.To.Hex()),
		Amount0In:   dbtypes.NewBigInt(data.Amount0In),
		Amount1In:   dbtypes.NewBigInt(data.Amount1In),
		Amount0Out:  dbtypes.NewBigInt(data.Amount0Out),
		Amount1Out:  dbtypes.NewBigInt(data.Amount1Out),
		TradePrice:  dbtypes.NewBigInt(nil),
	})
	if err != nil {
		return nil, err
	}
	return &CreatedDexTrade{Version: "uniswapv2", Trade: created}, nil
}

func (p *Processor) createV3Trade(ctx context.Context, chainID uint, event chainpubsub.ChainEventMessage) (*CreatedDexTrade, error) {
	if _, err := p.ensureV3Dex(ctx, chainID, event.Address); err != nil {
		return nil, err
	}
	var data chainentities.UniswapV3SwapEvent
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return nil, err
	}
	log.Println("Creating V3: ", chainID, event.Address, data.Sender.Hex(), data.Recipient.Hex())
	created, err := p.v3TradeSvc.Create(&dextradeentities.DexTradeUniswapV3{
		ChainID:      uint64(chainID),
		DexAddress:   strings.ToLower(event.Address),
		BlockNumber:  event.BlockNumber,
		PoolAddress:  strings.ToLower(event.Address),
		TxHash:       strings.ToLower(event.TxHash),
		Sender:       strings.ToLower(data.Sender.Hex()),
		Recipient:    strings.ToLower(data.Recipient.Hex()),
		Amount0:      dbtypes.NewBigInt(data.Amount0),
		Amount1:      dbtypes.NewBigInt(data.Amount1),
		SqrtPriceX96: dbtypes.NewBigInt(data.SqrtPriceX96),
		Liquidity:    dbtypes.NewBigInt(data.Liquidity),
		Tick:         data.Tick.Int64(),
		TradePrice:   dbtypes.NewBigInt(nil),
	})
	if err != nil {
		return nil, err
	}
	updated, err := p.v3TradeSvc.UpdateTradePrice(ctx, created.ID)
	if err != nil {
		return nil, err
	}
	return &CreatedDexTrade{Version: "uniswapv3", Trade: updated}, nil
}
