package dexactionsprocessor

import (
	"context"
	"encoding/json"
	"math/big"
	"strings"

	chainentities "github.com/alexkalak/whatever-system/src/modules/chain/entities"
	chainpubsub "github.com/alexkalak/whatever-system/src/modules/chain/pubsub"
	dexrepository "github.com/alexkalak/whatever-system/src/modules/exchanges/dex/repository"
	tokenservice "github.com/alexkalak/whatever-system/src/modules/tokens/service"
	dexactionentities "github.com/alexkalak/whatever-system/src/modules/trades/dexactions/entities"
	dexactionrepository "github.com/alexkalak/whatever-system/src/modules/trades/dexactions/repository"
	dexactionservice "github.com/alexkalak/whatever-system/src/modules/trades/dexactions/service"
	"github.com/alexkalak/whatever-system/src/shared/tools/dbtypes"
	"gorm.io/gorm"
)

type CreatedDexAction struct {
	Version string `json:"version"`
	Action  any    `json:"action"`
}

type Processor struct {
	v2ActionSvc dexactionservice.UniswapV2Service
	v3ActionSvc dexactionservice.UniswapV3Service
}

type Deps struct {
	DexRepo       dexrepository.DexRepository
	UniswapV3Repo dexrepository.UniswapV3Repository
	TokenService  tokenservice.TokenService
}

func New(db *gorm.DB, deps Deps) *Processor {
	_ = deps
	return &Processor{
		v2ActionSvc: dexactionservice.NewUniswapV2Service(dexactionrepository.NewUniswapV2Repository(db)),
		v3ActionSvc: dexactionservice.NewUniswapV3Service(dexactionrepository.NewUniswapV3Repository(db)),
	}
}

func (p *Processor) CreateActionsFromBlock(ctx context.Context, block chainpubsub.ChainBlockMessage) ([]CreatedDexAction, error) {
	actions := make([]CreatedDexAction, 0)
	for _, event := range block.Events {
		for range 3 {
			action, err := p.createActionFromEvent(ctx, block.ChainID, event)
			if err == nil {
				if action != nil {
					actions = append(actions, *action)
				}
				break
			}
		}
	}
	return actions, nil
}

func (p *Processor) createActionFromEvent(ctx context.Context, chainID uint, event chainpubsub.ChainEventMessage) (*CreatedDexAction, error) {
	_ = ctx
	switch event.Type {
	case chainentities.ChainSwapV2Event:
		return p.createV2SwapAction(chainID, event)
	case chainentities.ChainSwapV3Event:
		return p.createV3SwapAction(chainID, event)
	case chainentities.ChainMintV3Event:
		return p.createV3MintAction(chainID, event)
	case chainentities.ChainBurnV3Event:
		return p.createV3BurnAction(chainID, event)
	default:
		return nil, nil
	}
}

func (p *Processor) createV2SwapAction(chainID uint, event chainpubsub.ChainEventMessage) (*CreatedDexAction, error) {
	var data chainentities.UniswapV2SwapEvent
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return nil, err
	}
	amount0 := new(big.Int).Sub(data.Amount0Out, data.Amount0In)
	amount1 := new(big.Int).Sub(data.Amount1Out, data.Amount1In)
	created, err := p.v2ActionSvc.Create(&dexactionentities.DexActionUniswapV2{
		ChainID:      uint64(chainID),
		DexAddress:   strings.ToLower(event.Address),
		ActionType:   dexactionentities.ActionTypeSwap,
		BlockNumber:  event.BlockNumber,
		IndexInBlock: event.IndexInBlock,
		IndexInTx:    event.IndexInTx,
		PoolAddress:  strings.ToLower(event.Address),
		TxHash:       strings.ToLower(event.TxHash),
		Amount0:      dbtypes.NewBigInt(amount0),
		Amount1:      dbtypes.NewBigInt(amount1),
		Metadata: dbtypes.NewJSON(map[string]any{
			"sender":     strings.ToLower(data.Sender.Hex()),
			"recipient":  strings.ToLower(data.To.Hex()),
			"amount0In":  data.Amount0In.String(),
			"amount1In":  data.Amount1In.String(),
			"amount0Out": data.Amount0Out.String(),
			"amount1Out": data.Amount1Out.String(),
		}),
	})
	if err != nil {
		return nil, err
	}
	return &CreatedDexAction{Version: "uniswapv2", Action: created}, nil
}

func (p *Processor) createV3SwapAction(chainID uint, event chainpubsub.ChainEventMessage) (*CreatedDexAction, error) {
	var data chainentities.UniswapV3SwapEvent
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return nil, err
	}
	created, err := p.v3ActionSvc.Create(&dexactionentities.DexActionUniswapV3{
		ChainID:      uint64(chainID),
		DexAddress:   strings.ToLower(event.Address),
		ActionType:   dexactionentities.ActionTypeSwap,
		BlockNumber:  event.BlockNumber,
		IndexInBlock: event.IndexInBlock,
		IndexInTx:    event.IndexInTx,
		PoolAddress:  strings.ToLower(event.Address),
		TxHash:       strings.ToLower(event.TxHash),
		Amount0:      dbtypes.NewBigInt(data.Amount0),
		Amount1:      dbtypes.NewBigInt(data.Amount1),
		Metadata: dbtypes.NewJSON(map[string]any{
			"sender":       strings.ToLower(data.Sender.Hex()),
			"recipient":    strings.ToLower(data.Recipient.Hex()),
			"sqrtPriceX96": data.SqrtPriceX96.String(),
			"liquidity":    data.Liquidity.String(),
			"tick":         data.Tick.String(),
		}),
	})
	if err != nil {
		return nil, err
	}
	return &CreatedDexAction{Version: "uniswapv3", Action: created}, nil
}

func (p *Processor) createV3MintAction(chainID uint, event chainpubsub.ChainEventMessage) (*CreatedDexAction, error) {
	var data chainentities.UniswapV3MintEvent
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return nil, err
	}
	created, err := p.v3ActionSvc.Create(&dexactionentities.DexActionUniswapV3{
		ChainID:      uint64(chainID),
		DexAddress:   strings.ToLower(event.Address),
		ActionType:   dexactionentities.ActionTypeMint,
		BlockNumber:  event.BlockNumber,
		IndexInBlock: event.IndexInBlock,
		IndexInTx:    event.IndexInTx,
		PoolAddress:  strings.ToLower(event.Address),
		TxHash:       strings.ToLower(event.TxHash),
		Amount0:      dbtypes.NewBigInt(data.Amount0),
		Amount1:      dbtypes.NewBigInt(data.Amount1),
		Metadata: dbtypes.NewJSON(map[string]any{
			"sender":    strings.ToLower(data.Sender.Hex()),
			"owner":     strings.ToLower(data.Owner.Hex()),
			"liquidity": data.Amount.String(),
			"tickLower": data.TickLower,
			"tickUpper": data.TickUpper,
		}),
	})
	if err != nil {
		return nil, err
	}
	return &CreatedDexAction{Version: "uniswapv3", Action: created}, nil
}

func (p *Processor) createV3BurnAction(chainID uint, event chainpubsub.ChainEventMessage) (*CreatedDexAction, error) {
	var data chainentities.UniswapV3BurnEvent
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return nil, err
	}
	created, err := p.v3ActionSvc.Create(&dexactionentities.DexActionUniswapV3{
		ChainID:      uint64(chainID),
		DexAddress:   strings.ToLower(event.Address),
		ActionType:   dexactionentities.ActionTypeBurn,
		BlockNumber:  event.BlockNumber,
		IndexInBlock: event.IndexInBlock,
		IndexInTx:    event.IndexInTx,
		PoolAddress:  strings.ToLower(event.Address),
		TxHash:       strings.ToLower(event.TxHash),
		Amount0:      dbtypes.NewBigInt(data.Amount0),
		Amount1:      dbtypes.NewBigInt(data.Amount1),
		Metadata: dbtypes.NewJSON(map[string]any{
			"owner":     strings.ToLower(data.Owner.Hex()),
			"liquidity": data.Amount.String(),
			"tickLower": data.TickLower,
			"tickUpper": data.TickUpper,
		}),
	})
	if err != nil {
		return nil, err
	}
	return &CreatedDexAction{Version: "uniswapv3", Action: created}, nil
}
