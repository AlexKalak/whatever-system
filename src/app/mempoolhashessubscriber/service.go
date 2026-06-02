package mempoolhashessubscriber

import (
	"context"
	"log"
	"time"

	chainpubsub "github.com/alexkalak/whatever-system/src/modules/chain/pubsub"
	"github.com/alexkalak/whatever-system/src/modules/mempoolhashes/repository"
	mempoolhashesservice "github.com/alexkalak/whatever-system/src/modules/mempoolhashes/service"
	mempoolhashessubscriber "github.com/alexkalak/whatever-system/src/modules/mempoolhashes/subscriber"
	"github.com/alexkalak/whatever-system/src/modules/pubsub/kafka"
	"github.com/alexkalak/whatever-system/src/shared/tools/env"
	"gorm.io/gorm"
)

func Run(ctx context.Context, db *gorm.DB, brokers []string, envCfg env.Env) error {
	consumer := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers: brokers,
		Topic:   chainpubsub.ChainMempoolHashesTopic,
		GroupID: envCfg.KafkaMempoolConsumerGroup,
	})
	defer consumer.Close()

	repo := repository.NewMempoolHashRepository(db)
	mempoolHashService := mempoolhashesservice.NewMempoolHashService(repo)
	subscriber := mempoolhashessubscriber.NewChainMempoolHashSubscriber(consumer, mempoolhashessubscriber.ChainMempoolHashHandlerFunc(func(ctx context.Context, hash chainpubsub.ChainMempoolHashMessage) error {
		log.Printf("mempool hash received: chainId=%d txHash=%s ts=%s", hash.ChainID, hash.TxHash, hash.TS.Format(time.RFC3339Nano))
		return mempoolHashService.Save(ctx, hash.TxHash, hash.ChainID, hash.TS)
	}))

	return subscriber.Start(ctx)
}
