package main

import (
	"context"
	"log"

	"github.com/alexkalak/whatever-system/src/app/monolith"
	"github.com/alexkalak/whatever-system/src/shared/tools/env"
)

func main() {
	envCfg, err := env.GetEnv()
	if err != nil {
		log.Fatal(err)
	}

	if err := monolith.Run(context.Background(), envCfg); err != nil {
		log.Fatal(err)
	}
}
