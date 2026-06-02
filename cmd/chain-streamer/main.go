package main

import (
	"context"
	"log"

	"github.com/alexkalak/whatever-system/src/app/chainstreamer"
	"github.com/alexkalak/whatever-system/src/shared/tools/env"
	sharedkafka "github.com/alexkalak/whatever-system/src/shared/tools/kafka"
)

func main() {
	envCfg, err := env.GetEnv()
	if err != nil {
		log.Fatal(err)
	}

	brokers := sharedkafka.Brokers(envCfg.KafkaBrokers)
	if err := chainstreamer.Run(context.Background(), brokers, envCfg); err != nil {
		log.Fatal(err)
	}
}
