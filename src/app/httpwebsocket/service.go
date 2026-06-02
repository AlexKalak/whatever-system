package httpwebsocket

import (
	"context"
	"log"

	"github.com/alexkalak/whatever-system/src/modules/assets/controller"
	"github.com/alexkalak/whatever-system/src/modules/assets/repository"
	"github.com/alexkalak/whatever-system/src/modules/assets/service"
	dexcontroller "github.com/alexkalak/whatever-system/src/modules/exchanges/dex/controller"
	dexrepository "github.com/alexkalak/whatever-system/src/modules/exchanges/dex/repository"
	dexservice "github.com/alexkalak/whatever-system/src/modules/exchanges/dex/service"
	"github.com/alexkalak/whatever-system/src/modules/pubsub/kafka"
	tokencontroller "github.com/alexkalak/whatever-system/src/modules/tokens/controller"
	tokenrepository "github.com/alexkalak/whatever-system/src/modules/tokens/repository"
	tokenservice "github.com/alexkalak/whatever-system/src/modules/tokens/service"
	dexactioncontroller "github.com/alexkalak/whatever-system/src/modules/trades/dexactions/controller"
	actionpubsub "github.com/alexkalak/whatever-system/src/modules/trades/dexactions/pubsub"
	dexactionrepository "github.com/alexkalak/whatever-system/src/modules/trades/dexactions/repository"
	dexactionservice "github.com/alexkalak/whatever-system/src/modules/trades/dexactions/service"
	dexactionsubscriber "github.com/alexkalak/whatever-system/src/modules/trades/dexactions/subscriber"
	wsdexactions "github.com/alexkalak/whatever-system/src/modules/ws/dexactions"
	"github.com/alexkalak/whatever-system/src/shared/tools/env"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"gorm.io/gorm"
)

func Run(ctx context.Context, db *gorm.DB, brokers []string, envCfg env.Env) error {
	dexActionsHub := wsdexactions.NewHub()

	go func() {
		if err := consumeDexActionWebSocketEvents(ctx, brokers, envCfg, dexActionsHub); err != nil && ctx.Err() == nil {
			log.Printf("dex actions websocket event consumer stopped: %v", err)
		}
	}()

	return setupHTTPServer(db, dexActionsHub)
}

func consumeDexActionWebSocketEvents(ctx context.Context, brokers []string, envCfg env.Env, dexActionsHub *wsdexactions.Hub) error {
	consumer := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers: brokers,
		Topic:   actionpubsub.DexActionsCreatedTopic,
		GroupID: envCfg.KafkaDexActionsConsumerGroup + "-ws",
	})
	defer consumer.Close()

	subscriber := dexactionsubscriber.NewDexActionSubscriber(consumer, dexactionsubscriber.DexActionHandlerFunc(func(ctx context.Context, action actionpubsub.DexActionMessage) error {
		dexActionsHub.Broadcast(action)
		return nil
	}))
	return subscriber.Start(ctx)
}

func setupHTTPServer(db *gorm.DB, dexActionsHub *wsdexactions.Hub) error {
	assetRepo := repository.NewAssetRepository(db)
	assetService := service.NewAssetService(assetRepo)
	assetController := controller.NewAssetController(assetService)

	assetMappingRepo := repository.NewAssetMappingRepository(db)
	assetMappingService := service.NewAssetMappingService(assetMappingRepo)
	assetMappingController := controller.NewAssetMappingController(assetMappingService)

	dexRepo := dexrepository.NewDexRepository(db)
	dexService := dexservice.NewDexService(dexRepo)
	dexV2Repo := dexrepository.NewUniswapV2Repository(db)
	dexV3Repo := dexrepository.NewUniswapV3Repository(db)
	tokenRepo := tokenrepository.NewTokenRepository(db)
	tokenService := tokenservice.NewTokenService(tokenRepo, nil)
	dexController := dexcontroller.NewDexController(dexService, dexcontroller.DexControllerDeps{
		UniswapV2Repo: dexV2Repo,
		UniswapV3Repo: dexV3Repo,
	})
	tokenController := tokencontroller.NewTokenController(tokenService)

	dexActionV2Service := dexactionservice.NewUniswapV2Service(dexactionrepository.NewUniswapV2Repository(db))
	dexActionV3Service := dexactionservice.NewUniswapV3Service(dexactionrepository.NewUniswapV3Repository(db))
	dexActionController := dexactioncontroller.NewDexActionController(dexActionV2Service, dexActionV3Service, dexService)

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000,http://127.0.0.1:3000",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept",
	}))
	assetController.RegisterRoutes(app.Group("/assets"))
	assetMappingController.RegisterRoutes(app.Group("/asset-mappings"))
	dexController.RegisterRoutes(app.Group("/dexes"))
	tokenController.RegisterRoutes(app.Group("/tokens"))
	dexActionController.RegisterRoutes(app.Group("/dex-actions"))
	setupWebSocketRoutes(app, dexActionsHub)

	return app.Listen(":8080")
}

func setupWebSocketRoutes(app *fiber.App, dexActionsHub *wsdexactions.Hub) {
	dexActionsWSController := wsdexactions.NewController(dexActionsHub)
	dexActionsWSController.RegisterRoutes(app.Group("/ws"))
	log.Println("dex actions websocket route registered: /ws/dexactions?address=<address>&chainId=<chainId>")
}
