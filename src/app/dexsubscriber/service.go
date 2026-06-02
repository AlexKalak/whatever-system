package dexsubscriber

import (
	"context"

	chainpubsub "github.com/alexkalak/whatever-system/src/modules/chain/pubsub"
	chainservice "github.com/alexkalak/whatever-system/src/modules/chain/service"
	dexprocessor "github.com/alexkalak/whatever-system/src/modules/chainblocks/processor/dexprocessor"
	dexrepository "github.com/alexkalak/whatever-system/src/modules/exchanges/dex/repository"
	dexsubscriber "github.com/alexkalak/whatever-system/src/modules/exchanges/dex/subscriber"
	"github.com/alexkalak/whatever-system/src/modules/pubsub/kafka"
	tokenrepository "github.com/alexkalak/whatever-system/src/modules/tokens/repository"
	tokenservice "github.com/alexkalak/whatever-system/src/modules/tokens/service"
	"github.com/alexkalak/whatever-system/src/shared/tools/env"
	"gorm.io/gorm"
	"log"
)

func Run(ctx context.Context, db *gorm.DB, brokers []string, envCfg env.Env) error {
	consumer := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers: brokers,
		Topic:   chainpubsub.ChainBlocksTopic,
		GroupID: envCfg.KafkaDexConsumerGroup,
	})
	defer consumer.Close()

	chainData, err := chainservice.NewEVMChainDataService(
		map[uint]string{56: envCfg.BscRPCHTTPSURL, 1: envCfg.EthRPCHTTPSURL},
		map[uint]string{56: envCfg.BscMulticall3Address, 1: envCfg.EthMulticall3Address},
	)
	if err != nil {
		return err
	}
	dexRepo := dexrepository.NewDexRepository(db)
	v2DexRepo := dexrepository.NewUniswapV2Repository(db)
	v3DexRepo := dexrepository.NewUniswapV3Repository(db)
	tokenSvc := tokenservice.NewTokenService(tokenrepository.NewTokenRepository(db), chainData)
	processor := dexprocessor.New(db, dexprocessor.Deps{
		ChainData:     chainData,
		DexRepo:       dexRepo,
		UniswapV2Repo: v2DexRepo,
		UniswapV3Repo: v3DexRepo,
		TokenService:  tokenSvc,
	})

	subscriber := dexsubscriber.NewChainBlockSubscriber(consumer, dexsubscriber.ChainBlockHandlerFunc(func(ctx context.Context, block chainpubsub.ChainBlockMessage) error {
		log.Printf("%d dex chain block received: chainId=%d events=%d", block.BlockNumber, block.ChainID, len(block.Events))
		return processor.ProcessEvent(ctx, block)
	}))

	return subscriber.Start(ctx)
}
