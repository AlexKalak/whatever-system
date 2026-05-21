package main

import (
	"log"
	"sync"

	"github.com/alexkalak/whatever-system/src/infrastructure/postgresdb"
	"github.com/alexkalak/whatever-system/src/modules/assets/controller"
	"github.com/alexkalak/whatever-system/src/modules/assets/repository"
	"github.com/alexkalak/whatever-system/src/modules/assets/service"
	"github.com/alexkalak/whatever-system/src/modules/eventbus"
	dexcontroller "github.com/alexkalak/whatever-system/src/modules/exchanges/dex/controller"
	dexlistener "github.com/alexkalak/whatever-system/src/modules/exchanges/dex/listener"
	dexrepository "github.com/alexkalak/whatever-system/src/modules/exchanges/dex/repository"
	dexservice "github.com/alexkalak/whatever-system/src/modules/exchanges/dex/service"
	"github.com/alexkalak/whatever-system/src/shared/tools/env"
	"github.com/gofiber/fiber/v2"
)

func main() {
	envCfg, err := env.GetEnv()
	if err != nil {
		log.Fatal(err)
	}

	db := postgresdb.InitDB(envCfg)

	wg := sync.WaitGroup{}

	wg.Go(func() {
		assetRepo := repository.NewAssetRepository(db)
		assetService := service.NewAssetService(assetRepo)
		assetController := controller.NewAssetController(assetService)

		assetMappingRepo := repository.NewAssetMappingRepository(db)
		assetMappingService := service.NewAssetMappingService(assetMappingRepo)
		assetMappingController := controller.NewAssetMappingController(assetMappingService)

		dexRepo := dexrepository.NewDexRepository(db)
		dexService := dexservice.NewDexService(dexRepo)
		dexController := dexcontroller.NewDexController(dexService)

		bus := eventbus.NewEventBus()
		dexChainEventsService := dexservice.NewDexChainEventsService(dexRepo)
		dexlistener.RegisterChainSwapV3Listener(bus, dexChainEventsService)

		app := fiber.New()
		assetController.RegisterRoutes(app.Group("/assets"))
		assetMappingController.RegisterRoutes(app.Group("/asset-mappings"))
		dexController.RegisterRoutes(app.Group("/dexes"))
		if err := app.Listen(":8080"); err != nil {
			log.Fatal(err)
		}
	})

	wg.Go(func() {
	})

	wg.Wait()
}
