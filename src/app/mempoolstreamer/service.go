package mempoolstreamer

import (
	"context"
	"fmt"
	"log"
	"time"

	chainpublisher "github.com/alexkalak/whatever-system/src/modules/chain/publisher"
	chainservice "github.com/alexkalak/whatever-system/src/modules/chain/service"
	"github.com/alexkalak/whatever-system/src/modules/pubsub/kafka"
	"github.com/alexkalak/whatever-system/src/shared/tools/env"
)

func Run(ctx context.Context, brokers []string, envCfg env.Env) error {
	kafkaPublisher := kafka.NewPublisher(kafka.PublisherConfig{Brokers: brokers})
	defer kafkaPublisher.Close()
	mempoolPublisher := chainpublisher.NewChainMempoolPublisher(kafkaPublisher)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		streamCtx, cancelStream := context.WithCancel(ctx)
		streamer, err := chainservice.NewChainMempoolStreamer(streamCtx, chainservice.ChainMempoolStreamerConfig{
			ChainID:            1,
			WsRPCURL:           envCfg.EthRPCWsURL,
			SubscriptionMethod: "alchemy_pendingTransactions",
			SubscriptionParams: []any{
				map[string]any{
					"hashesOnly": false,
				},
			},
		})
		if err != nil {
			cancelStream()
			log.Printf("chain mempool streamer init failed: %v", err)
			time.Sleep(time.Second)
			continue
		}

		streams, err := streamer.Start(streamCtx)
		if err != nil {
			cancelStream()
			log.Printf("chain mempool streamer start failed: %v", err)
			time.Sleep(time.Second)
			continue
		}

		errCh := make(chan error, 2)
		go func() {
			for {
				select {
				case <-streamCtx.Done():
					errCh <- nil
					return
				case hash, ok := <-streams.Hashes:
					if !ok {
						errCh <- nil
						return
					}
					if err := mempoolPublisher.PublishHash(streamCtx, hash); err != nil {
						errCh <- fmt.Errorf("publish mempool hash: %w", err)
						cancelStream()
						return
					}
				}
			}
		}()

		go func() {
			for {
				select {
				case <-streamCtx.Done():
					errCh <- nil
					return
				case event, ok := <-streams.Transactions:
					if !ok {
						errCh <- nil
						return
					}
					if err := mempoolPublisher.Publish(streamCtx, event); err != nil {
						errCh <- fmt.Errorf("publish mempool transaction: %w", err)
						cancelStream()
						return
					}
				}
			}
		}()

		var publishErr error
		for range 2 {
			if err := <-errCh; err != nil && publishErr == nil {
				publishErr = err
				cancelStream()
			}
		}
		cancelStream()
		if publishErr != nil {
			return publishErr
		}

		log.Println("chain mempool streamer stopped, restarting")
		time.Sleep(time.Second)
	}
}
