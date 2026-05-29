package dextrades

import (
	"encoding/json"
	"log"
	"strings"
	"sync"

	chainblockprocessor "github.com/alexkalak/whatever-system/src/modules/chainblocks/processor"
	tradeentities "github.com/alexkalak/whatever-system/src/modules/trades/dextrades/entities"
)

type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[*Client]struct{})}
}

func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	h.clients[client] = struct{}{}
	h.mu.Unlock()
}

func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.send)
	}
	h.mu.Unlock()
}

func (h *Hub) Broadcast(payload any) {
	message, err := json.Marshal(payload)
	if err != nil {
		log.Printf("dex trades ws marshal failed: %v", err)
		return
	}

	h.mu.RLock()
	clients := make([]*Client, 0, len(h.clients))
	for client := range h.clients {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		if !matchesClientFilter(client, payload) {
			continue
		}

		select {
		case client.send <- message:
		default:
			h.Unregister(client)
		}
	}
}

func matchesClientFilter(client *Client, payload any) bool {
	switch trade := payload.(type) {
	case chainblockprocessor.CreatedDexTrade:
		return matchesClientFilter(client, trade.Trade)
	case *chainblockprocessor.CreatedDexTrade:
		if trade == nil {
			return false
		}
		return matchesClientFilter(client, trade.Trade)
	case tradeentities.DexTradeUniswapV2:
		return matchesTrade(client, trade.ChainID, trade.DexAddress, trade.PoolAddress, trade.Sender, trade.Recipient)
	case *tradeentities.DexTradeUniswapV2:
		if trade == nil {
			return false
		}
		return matchesTrade(client, trade.ChainID, trade.DexAddress, trade.PoolAddress, trade.Sender, trade.Recipient)
	case tradeentities.DexTradeUniswapV3:
		return matchesTrade(client, trade.ChainID, trade.DexAddress, trade.PoolAddress, trade.Sender, trade.Recipient)
	case *tradeentities.DexTradeUniswapV3:
		if trade == nil {
			return false
		}
		return matchesTrade(client, trade.ChainID, trade.DexAddress, trade.PoolAddress, trade.Sender, trade.Recipient)
	default:
		return false
	}
}

func matchesTrade(client *Client, chainID uint64, addresses ...string) bool {
	if client.chainID != chainID {
		return false
	}

	for _, address := range addresses {
		if strings.EqualFold(strings.TrimSpace(address), client.address) {
			return true
		}
	}
	return false
}
