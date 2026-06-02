package dexactionssubscriber

import (
	"context"
	"fmt"
	"log"

	chainpubsub "github.com/alexkalak/whatever-system/src/modules/chain/pubsub"
	chainservice "github.com/alexkalak/whatever-system/src/modules/chain/service"
	dexactionsprocessor "github.com/alexkalak/whatever-system/src/modules/chainblocks/processor/dexactionsprocessor"
	dexrepository "github.com/alexkalak/whatever-system/src/modules/exchanges/dex/repository"
	"github.com/alexkalak/whatever-system/src/modules/pubsub/kafka"
	tokenrepository "github.com/alexkalak/whatever-system/src/modules/tokens/repository"
	tokenservice "github.com/alexkalak/whatever-system/src/modules/tokens/service"
	dexactionpublisher "github.com/alexkalak/whatever-system/src/modules/trades/dexactions/publisher"
	dexactionsubscriber "github.com/alexkalak/whatever-system/src/modules/trades/dexactions/subscriber"
	"github.com/alexkalak/whatever-system/src/shared/tools/env"
	"gorm.io/gorm"
)

func Run(ctx context.Context, db *gorm.DB, brokers []string, envCfg env.Env) error {
	consumer := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers: brokers,
		Topic:   chainpubsub.ChainBlocksTopic,
		GroupID: envCfg.KafkaDexActionsConsumerGroup,
	})
	defer consumer.Close()

	kafkaPublisher := kafka.NewPublisher(kafka.PublisherConfig{Brokers: brokers})
	defer kafkaPublisher.Close()
	actionPublisher := dexactionpublisher.NewDexActionPublisher(kafkaPublisher)

	chainData, err := chainservice.NewEVMChainDataService(
		map[uint]string{56: envCfg.BscRPCHTTPSURL, 1: envCfg.EthRPCHTTPSURL},
		map[uint]string{56: envCfg.BscMulticall3Address, 1: envCfg.EthMulticall3Address},
	)
	if err != nil {
		return err
	}
	dexRepo := dexrepository.NewDexRepository(db)
	v3DexRepo := dexrepository.NewUniswapV3Repository(db)
	tokenSvc := tokenservice.NewTokenService(tokenrepository.NewTokenRepository(db), chainData)
	processor := dexactionsprocessor.New(db, dexactionsprocessor.Deps{
		DexRepo:       dexRepo,
		UniswapV3Repo: v3DexRepo,
		TokenService:  tokenSvc,
	})

	subscriber := dexactionsubscriber.NewChainBlockSubscriber(consumer, dexactionsubscriber.ChainBlockHandlerFunc(func(ctx context.Context, block chainpubsub.ChainBlockMessage) error {
		log.Printf("%d dex actions chain block received: chainId=%d events=%d", block.BlockNumber, block.ChainID, len(block.Events))
		actions, err := processor.CreateActionsFromBlock(ctx, block)
		if err != nil {
			return err
		}
		for _, action := range actions {
			if err := actionPublisher.Publish(ctx, action); err != nil {
				return err
			}
		}
		return nil
	}))
	fmt.Println("Dex actions subscriber started")
	return subscriber.Start(ctx)
}
