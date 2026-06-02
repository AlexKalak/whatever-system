package dexactions

import (
	"encoding/json"
	"log"
	"strings"
	"sync"

	dexactionsprocessor "github.com/alexkalak/whatever-system/src/modules/chainblocks/processor/dexactionsprocessor"
	actionentities "github.com/alexkalak/whatever-system/src/modules/trades/dexactions/entities"
	actionpubsub "github.com/alexkalak/whatever-system/src/modules/trades/dexactions/pubsub"
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
		log.Printf("dex actions ws marshal failed: %v", err)
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
	switch action := payload.(type) {
	case dexactionsprocessor.CreatedDexAction:
		return matchesClientFilter(client, action.Action)
	case *dexactionsprocessor.CreatedDexAction:
		if action == nil {
			return false
		}
		return matchesClientFilter(client, action.Action)
	case actionpubsub.DexActionMessage:
		return matchesAction(client, action.ChainID, action.DexAddress, action.PoolAddress)
	case *actionpubsub.DexActionMessage:
		if action == nil {
			return false
		}
		return matchesAction(client, action.ChainID, action.DexAddress, action.PoolAddress)
	case actionentities.DexActionUniswapV2:
		return matchesAction(client, action.ChainID, action.DexAddress, action.PoolAddress)
	case *actionentities.DexActionUniswapV2:
		if action == nil {
			return false
		}
		return matchesAction(client, action.ChainID, action.DexAddress, action.PoolAddress)
	case actionentities.DexActionUniswapV3:
		return matchesAction(client, action.ChainID, action.DexAddress, action.PoolAddress)
	case *actionentities.DexActionUniswapV3:
		if action == nil {
			return false
		}
		return matchesAction(client, action.ChainID, action.DexAddress, action.PoolAddress)
	default:
		return false
	}
}

func matchesAction(client *Client, chainID uint64, addresses ...string) bool {
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
