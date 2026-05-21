package chain

import (
	"context"

	chainentities "github.com/alexkalak/whatever-system/src/modules/chain/entities"
	chainservice "github.com/alexkalak/whatever-system/src/modules/chain/service"
	"github.com/alexkalak/whatever-system/src/modules/eventbus"
)

type EventsAdapter struct {
	sender eventbus.Sender
}

func NewEventsAdapter(sender eventbus.Sender) *EventsAdapter {
	return &EventsAdapter{sender: sender}
}

func (a *EventsAdapter) Forward(ctx context.Context, in <-chan chainservice.ChainEventChannelEntity) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case chainEvent, ok := <-in:
			if !ok {
				return nil
			}

			if err := a.publish(ctx, chainEvent); err != nil {
				return err
			}
		}
	}
}

func (a *EventsAdapter) publish(ctx context.Context, chainEvent chainservice.ChainEventChannelEntity) error {
	event := chainEvent.Event
	ts := chainEvent.TS

	switch data := event.Data.(type) {
	case chainentities.UniswapV3SwapEvent:
		return eventbus.Publish(ctx, a.sender, SwapV3Topic, mapSwapV3Payload(chainEvent.ChainID, "uniswap_v3", event, data.Sender.String(), data.Recipient.String(), data.Amount0.String(), data.Amount1.String()), ts)
	case chainentities.PancakeswapV3SwapEvent:
		return eventbus.Publish(ctx, a.sender, SwapV3Topic, mapSwapV3Payload(chainEvent.ChainID, "pancakeswap_v3", event, data.Sender.String(), data.Recipient.String(), data.Amount0.String(), data.Amount1.String()), ts)
	case chainentities.SushiswapV3SwapEvent:
		return eventbus.Publish(ctx, a.sender, SwapV3Topic, mapSwapV3Payload(chainEvent.ChainID, "sushiswap_v3", event, data.Sender.String(), data.Recipient.String(), data.Amount0.String(), data.Amount1.String()), ts)
	default:
		return eventbus.Publish(ctx, a.sender, UnknownTopic, UnknownPayload{
			EventType:   event.Type,
			BlockNumber: event.BlockNumber,
			Address:     event.Address,
			TxHash:      event.TxHash,
		}, ts)
	}
}

func mapSwapV3Payload(chainID uint, dex string, event chainentities.ChainEvent, sender string, recipient string, amount0 string, amount1 string) SwapV3Payload {
	return SwapV3Payload{
		ChainID:     chainID,
		Dex:         dex,
		BlockNumber: event.BlockNumber,
		PoolAddress: event.Address,
		TxHash:      event.TxHash,
		Sender:      sender,
		Recipient:   recipient,
		Amount0:     amount0,
		Amount1:     amount1,
	}
}
