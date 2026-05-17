package main

import (
	"log"

	"github.com/alexkalak/whatever-system/src/infrastructure/postgresdb"
	"github.com/alexkalak/whatever-system/src/modules/assets/controller"
	"github.com/alexkalak/whatever-system/src/modules/assets/repository"
	"github.com/alexkalak/whatever-system/src/modules/assets/service"
	"github.com/alexkalak/whatever-system/src/shared/tools/env"
	"github.com/gofiber/fiber/v2"
)

func main() {
	envCfg, err := env.GetEnv()
	if err != nil {
		log.Fatal(err)
	}

	db := postgresdb.InitDB(envCfg)

	assetRepo := repository.NewAssetRepository(db)
	assetService := service.NewAssetService(assetRepo)
	assetController := controller.NewAssetController(assetService)

	assetMappingRepo := repository.NewAssetMappingRepository(db)
	assetMappingService := service.NewAssetMappingService(assetMappingRepo)
	assetMappingController := controller.NewAssetMappingController(assetMappingService)

	app := fiber.New()
	assetController.RegisterRoutes(app.Group("/assets"))
	assetMappingController.RegisterRoutes(app.Group("/asset-mappings"))

	if err := app.Listen(":8080"); err != nil {
		log.Fatal(err)
	}
}
