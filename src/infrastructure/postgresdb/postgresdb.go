package postgresdb

import (
	"fmt"
	"log"
	"os"
	"time"

	assetentities "github.com/alexkalak/whatever-system/src/modules/assets/entities"
	dexentities "github.com/alexkalak/whatever-system/src/modules/exchanges/dex/entities"
	tokenentities "github.com/alexkalak/whatever-system/src/modules/tokens/entities"
	dextradeentities "github.com/alexkalak/whatever-system/src/modules/trades/dextrades/entities"
	"github.com/alexkalak/whatever-system/src/shared/tools/env"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB(env env.Env) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		env.PostgresHost,
		env.PostgresUser,
		env.PostgresPassword,
		env.PostgresDBName,
		env.PostgresPort,
		"disable",
	)

	dbLogger := initDBErrorLogger()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: dbLogger})
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}

	if err := db.AutoMigrate(
		&assetentities.Asset{},
		&assetentities.AssetMapping{},
		&dexentities.Dex{},
		&dexentities.UniswapV2Dex{},
		&dexentities.UniswapV3Dex{},
		&tokenentities.Token{},
		&dextradeentities.DexTradeUniswapV2{},
		&dextradeentities.DexTradeUniswapV3{},
	); err != nil {
		log.Fatal("failed to migrate database:", err)
	}

	fmt.Println("Database initialized")

	return db
}

func initDBErrorLogger() logger.Interface {
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Println("failed to create logs directory:", err)
		return logger.Default.LogMode(logger.Error)
	}

	file, err := os.OpenFile("logs/database_errors.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Println("failed to open database error log file:", err)
		return logger.Default.LogMode(logger.Error)
	}

	return logger.New(
		log.New(file, "", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Error,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)
}
