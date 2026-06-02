package main

import (
	"context"
	"log"

	"github.com/alexkalak/whatever-system/src/app/mempoolstreamer"
	"github.com/alexkalak/whatever-system/src/shared/tools/env"
	sharedkafka "github.com/alexkalak/whatever-system/src/shared/tools/kafka"
)

func main() {
	envCfg, err := env.GetEnv()
	if err != nil {
		log.Fatal(err)
	}

	brokers := sharedkafka.Brokers(envCfg.KafkaBrokers)
	ctx := context.Background()

	err = mempoolstreamer.Run(ctx, brokers, envCfg)
	if err != nil {
		log.Fatal(err)
	}
}
