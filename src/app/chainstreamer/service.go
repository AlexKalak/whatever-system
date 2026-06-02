package chainstreamer

import (
	"context"
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
	blockPublisher := chainpublisher.NewChainBlockPublisher(kafkaPublisher)

	var lastPublishedBlock uint64
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		fromBlock := uint64(0)
		if lastPublishedBlock > 0 {
			fromBlock = lastPublishedBlock + 1
		}

		streamer, err := chainservice.NewChainLogsStreamer(ctx, chainservice.ChainLogsStreamerConfig{
			ChainID:    1,
			WsRPCURL:   envCfg.EthRPCWsURL,
			HTTPRPCURL: envCfg.EthRPCHTTPSURL,
		})
		if err != nil {
			log.Printf("chain logs streamer init failed: %v", err)
			time.Sleep(time.Second)
			continue
		}

		blocksCh, err := streamer.Start(ctx, fromBlock, nil)
		if err != nil {
			log.Printf("chain logs streamer start failed from block %d: %v", fromBlock, err)
			time.Sleep(time.Second)
			continue
		}

		for block := range blocksCh {
			if err := blockPublisher.Publish(ctx, block); err != nil {
				return err
			}
			lastPublishedBlock = block.BlockNumber
		}

		log.Printf("chain logs streamer stopped, restarting from block %d", lastPublishedBlock)
		time.Sleep(time.Second)
	}
}
