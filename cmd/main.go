package main

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/alexkalak/whatever-system/src/infrastructure/postgresdb"
	"github.com/alexkalak/whatever-system/src/modules/assets/controller"
	"github.com/alexkalak/whatever-system/src/modules/assets/repository"
	"github.com/alexkalak/whatever-system/src/modules/assets/service"
	chainpublisher "github.com/alexkalak/whatever-system/src/modules/chain/publisher"
	chainpubsub "github.com/alexkalak/whatever-system/src/modules/chain/pubsub"
	chainservice "github.com/alexkalak/whatever-system/src/modules/chain/service"
	chainsubscriber "github.com/alexkalak/whatever-system/src/modules/chain/subscriber"
	chainblockprocessor "github.com/alexkalak/whatever-system/src/modules/chainblocks/processor"
	dexcontroller "github.com/alexkalak/whatever-system/src/modules/exchanges/dex/controller"
	dexrepository "github.com/alexkalak/whatever-system/src/modules/exchanges/dex/repository"
	dexservice "github.com/alexkalak/whatever-system/src/modules/exchanges/dex/service"
	dexsubscriber "github.com/alexkalak/whatever-system/src/modules/exchanges/dex/subscriber"
	"github.com/alexkalak/whatever-system/src/modules/pubsub/kafka"
	dextradessubscriber "github.com/alexkalak/whatever-system/src/modules/trades/dextrades/subscriber"
	wsdextrades "github.com/alexkalak/whatever-system/src/modules/ws/dextrades"
	"github.com/alexkalak/whatever-system/src/shared/tools/env"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func main() {
	envCfg, err := env.GetEnv()
	if err != nil {
		log.Fatal(err)
	}

	db := postgresdb.InitDB(envCfg)
	ctx := context.Background()
	brokers := kafkaBrokers(envCfg.KafkaBrokers)

	dexTradesHub := wsdextrades.NewHub()

	wg := sync.WaitGroup{}
	wg.Go(func() {
		if err := setupHTTPServer(db, dexTradesHub); err != nil {
			log.Fatal(err)
		}
	})
	wg.Go(func() {
		if err := setupChainBlockPublisher(ctx, brokers, envCfg); err != nil {
			log.Fatal(err)
		}
	})

	wg.Go(func() {
		if err := setupDexChainBlockSubscriber(ctx, db, brokers, envCfg); err != nil {
			log.Fatal(err)
		}
	})
	wg.Go(func() {
		if err := setupDexTradesChainBlockSubscriber(ctx, db, brokers, envCfg, dexTradesHub); err != nil {
			log.Fatal(err)
		}
	})

	wg.Wait()
}

func kafkaBrokers(raw string) []string {
	parts := strings.Split(raw, ",")
	brokers := make([]string, 0, len(parts))
	for _, part := range parts {
		broker := strings.TrimSpace(part)
		if broker != "" {
			brokers = append(brokers, broker)
		}
	}
	return brokers
}

func setupChainBlockPublisher(ctx context.Context, brokers []string, envCfg env.Env) error {
	kafkaPublisher := kafka.NewPublisher(kafka.PublisherConfig{Brokers: brokers})
	blockPublisher := chainpublisher.NewChainBlockPublisher(kafkaPublisher)

	var lastPublishedBlock uint64
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		fromBlock := uint64(0)
		if lastPublishedBlock > 0 {
			fromBlock = lastPublishedBlock + 1
		}

		streamer, err := chainservice.NewChainLogsStreamer(ctx, chainservice.ChainLogsStreamerConfig{
			ChainID:    56,
			WsRPCURL:   envCfg.BscRPCWsURL,
			HTTPRPCURL: envCfg.BscRPCHTTPSURL,
		})
		if err != nil {
			log.Printf("chain logs streamer init failed: %v", err)
			time.Sleep(time.Second)
			continue
		}

		blocksCh, err := streamer.Start(ctx, fromBlock, nil)
		if err != nil {
			log.Printf("chain logs streamer start failed from block %d: %v", fromBlock, err)
			time.Sleep(time.Second)
			continue
		}

		for block := range blocksCh {
			if err := blockPublisher.Publish(ctx, block); err != nil {
				return err
			}
			lastPublishedBlock = block.BlockNumber
		}

		log.Printf("chain logs streamer stopped, restarting from block %d", fromBlock)
		log.Printf("1s zzz...")
		time.Sleep(time.Second)
	}
}

func setupChainMempoolPublisher(ctx context.Context, brokers []string, envCfg env.Env) error {
	kafkaPublisher := kafka.NewPublisher(kafka.PublisherConfig{Brokers: brokers})
	mempoolPublisher := chainpublisher.NewChainMempoolPublisher(kafkaPublisher)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		streamer, err := chainservice.NewChainMempoolStreamer(ctx, chainservice.ChainMempoolStreamerConfig{
			ChainID:  56,
			WsRPCURL: envCfg.BscRPCWsURL,
		})
		if err != nil {
			log.Printf("chain mempool streamer init failed: %v", err)
			time.Sleep(time.Second)
			continue
		}

		txsCh, err := streamer.Start(ctx)
		if err != nil {
			log.Printf("chain mempool streamer start failed: %v", err)
			time.Sleep(time.Second)
			continue
		}

		for tx := range txsCh {
			if err := mempoolPublisher.Publish(ctx, tx); err != nil {
				return err
			}
		}

		log.Printf("chain mempool streamer stopped, restarting")
		log.Printf("1s zzz...")
		time.Sleep(time.Second)
	}
}

func setupChainMempoolSubscriber(ctx context.Context, brokers []string, envCfg env.Env) error {
	consumer := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers: brokers,
		Topic:   chainpubsub.ChainMempoolEventsTopic,
		GroupID: envCfg.KafkaMempoolConsumerGroup,
	})

	subscriber := chainsubscriber.NewChainMempoolSubscriber(consumer, chainsubscriber.ChainMempoolHandlerFunc(func(ctx context.Context, tx chainpubsub.ChainMempoolEventMessage) error {
		log.Printf("mempool tx received: chainId=%d txHash=%s from=%s to=%s value=%s method=%s signature=%s args=%v", tx.ChainID, tx.TxHash, tx.From, tx.To, tx.Value, tx.CallData.Method, tx.CallData.Signature, tx.CallData.Args)
		return nil
	}))

	return subscriber.Start(ctx)
}

func setupDexChainBlockSubscriber(ctx context.Context, db *gorm.DB, brokers []string, envCfg env.Env) error {
	consumer := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers: brokers,
		Topic:   chainpubsub.ChainBlocksTopic,
		GroupID: envCfg.KafkaDexConsumerGroup,
	})
	processor, err := chainblockprocessor.New(db, map[uint]string{56: envCfg.BscRPCHTTPSURL})
	if err != nil {
		return err
	}

	subscriber := dexsubscriber.NewChainBlockSubscriber(consumer, dexsubscriber.ChainBlockHandlerFunc(func(ctx context.Context, block chainpubsub.ChainBlockMessage) error {
		log.Printf("dex chain block received: chainId=%d block=%d events=%d", block.ChainID, block.BlockNumber, len(block.Events))
		return processor.Process(ctx, block)
	}))

	return subscriber.Start(ctx)
}

func setupDexTradesChainBlockSubscriber(ctx context.Context, db *gorm.DB, brokers []string, envCfg env.Env, dexTradesHub *wsdextrades.Hub) error {
	consumer := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers: brokers,
		Topic:   chainpubsub.ChainBlocksTopic,
		GroupID: envCfg.KafkaDexTradesConsumerGroup,
	})
	processor, err := chainblockprocessor.New(db, map[uint]string{56: envCfg.BscRPCHTTPSURL})
	if err != nil {
		return err
	}

	subscriber := dextradessubscriber.NewChainBlockSubscriber(consumer, dextradessubscriber.ChainBlockHandlerFunc(func(ctx context.Context, block chainpubsub.ChainBlockMessage) error {
		log.Printf("dex trades chain block received: chainId=%d block=%d events=%d", block.ChainID, block.BlockNumber, len(block.Events))
		trades, err := processor.CreateTradesFromBlock(ctx, block)
		if err != nil {
			return err
		}
		for _, trade := range trades {
			dexTradesHub.Broadcast(trade)
		}
		return nil
	}))

	return subscriber.Start(ctx)
}

func setupHTTPServer(db *gorm.DB, dexTradesHub *wsdextrades.Hub) error {
	assetRepo := repository.NewAssetRepository(db)
	assetService := service.NewAssetService(assetRepo)
	assetController := controller.NewAssetController(assetService)

	assetMappingRepo := repository.NewAssetMappingRepository(db)
	assetMappingService := service.NewAssetMappingService(assetMappingRepo)
	assetMappingController := controller.NewAssetMappingController(assetMappingService)

	dexRepo := dexrepository.NewDexRepository(db)
	dexService := dexservice.NewDexService(dexRepo)
	dexController := dexcontroller.NewDexController(dexService)

	app := fiber.New()
	assetController.RegisterRoutes(app.Group("/assets"))
	assetMappingController.RegisterRoutes(app.Group("/asset-mappings"))
	dexController.RegisterRoutes(app.Group("/dexes"))
	setupWebSocketRoutes(app, dexTradesHub)

	return app.Listen(":8080")
}

func setupWebSocketRoutes(app *fiber.App, dexTradesHub *wsdextrades.Hub) {
	dexTradesWSController := wsdextrades.NewController(dexTradesHub)
	dexTradesWSController.RegisterRoutes(app.Group("/ws"))
	log.Println("dex trades websocket route registered: /ws/dextrades?address=<address>&chainId=<chainId>")
}
