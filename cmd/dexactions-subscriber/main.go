package main

import (
	"context"
	"log"

	"github.com/alexkalak/whatever-system/src/app/dexactionssubscriber"
	"github.com/alexkalak/whatever-system/src/infrastructure/postgresdb"
	"github.com/alexkalak/whatever-system/src/shared/tools/env"
	sharedkafka "github.com/alexkalak/whatever-system/src/shared/tools/kafka"
)

func main() {
	envCfg, err := env.GetEnv()
	if err != nil {
		log.Fatal(err)
	}

	db := postgresdb.InitDB(envCfg)
	brokers := sharedkafka.Brokers(envCfg.KafkaBrokers)
	if err := dexactionssubscriber.Run(context.Background(), db, brokers, envCfg); err != nil {
		log.Fatal(err)
	}
}
