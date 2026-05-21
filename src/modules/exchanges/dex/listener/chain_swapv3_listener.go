package listener

import (
	"context"

	chaineventbus "github.com/alexkalak/whatever-system/src/modules/eventbus/chain"
	"github.com/alexkalak/whatever-system/src/modules/eventbus"
	dexservice "github.com/alexkalak/whatever-system/src/modules/exchanges/dex/service"
)

func RegisterChainSwapV3Listener(bus *eventbus.EventBus, service dexservice.DexChainEventsService) {
	eventbus.SubscribeTopic(bus, chaineventbus.SwapV3Topic, func(ctx context.Context, event eventbus.Event[chaineventbus.SwapV3Payload]) {
		_ = service.EnsureDexFromSwapV3(ctx, event.Payload)
	})
}
