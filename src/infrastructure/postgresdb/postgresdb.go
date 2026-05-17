package postgresdb

import (
	"fmt"
	"log"

	assetentities "github.com/alexkalak/whatever-system/src/modules/assets/entities"
	"github.com/alexkalak/whatever-system/src/shared/tools/env"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}

	db.AutoMigrate(
		&assetentities.Asset{},
		&assetentities.AssetMapping{},
	)

	fmt.Println("Database initialized")

	return db
}
