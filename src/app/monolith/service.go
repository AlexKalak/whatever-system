package monolith

import (
	"context"
	"log"
	"sync"

	"github.com/alexkalak/whatever-system/src/app/chainstreamer"
	"github.com/alexkalak/whatever-system/src/app/dexactionssubscriber"
	"github.com/alexkalak/whatever-system/src/app/dexsubscriber"
	"github.com/alexkalak/whatever-system/src/app/httpwebsocket"
	// "github.com/alexkalak/whatever-system/src/app/mempoolhashessubscriber"
	// "github.com/alexkalak/whatever-system/src/app/mempoolstreamer"
	"github.com/alexkalak/whatever-system/src/infrastructure/postgresdb"
	"github.com/alexkalak/whatever-system/src/shared/tools/env"
	sharedkafka "github.com/alexkalak/whatever-system/src/shared/tools/kafka"
)

func Run(ctx context.Context, envCfg env.Env) error {
	db := postgresdb.InitDB(envCfg)
	brokers := sharedkafka.Brokers(envCfg.KafkaBrokers)

	wg := sync.WaitGroup{}
	errCh := make(chan error, 4)
	start := func(name string, fn func() error) {
		wg.Go(func() {
			log.Printf("starting %s", name)
			if err := fn(); err != nil {
				errCh <- err
			}
		})
	}

	start("http-websocket", func() error { return httpwebsocket.Run(ctx, db, brokers, envCfg) })

	start("chain-streamer", func() error { return chainstreamer.Run(ctx, brokers, envCfg) })

	// start("mempool-streamer", func() error { return mempoolstreamer.Run(ctx, brokers, envCfg) })
	// start("mempool-hashes-subscriber", func() error { return mempoolhashessubscriber.Run(ctx, db, brokers, envCfg) })

	start("dex-subscriber", func() error { return dexsubscriber.Run(ctx, db, brokers, envCfg) })
	start("dexactions-subscriber", func() error { return dexactionssubscriber.Run(ctx, db, brokers, envCfg) })

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		wg.Wait()
		return ctx.Err()
	}
}
