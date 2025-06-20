package ws

import (
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/labstack/gommon/log"
	"net/http"
	"sync"
	"time"
)

// Hub manage all client connections
type Hub struct {
	clients       map[*Client]bool
	register      chan *Client
	unregister    chan *Client
	broadcast     chan []byte
	subscriptions map[SubscriptionKey]map[*Client]bool
	mu            sync.RWMutex
}

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// check origin in prod env
		return true
	},
}

func NewHub() *Hub {
	return &Hub{
		clients:       make(map[*Client]bool),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		broadcast:     make(chan []byte),
		subscriptions: make(map[SubscriptionKey]map[*Client]bool),
	}
}

func (h *Hub) GetSubscriptionKeys() []SubscriptionKey {
	h.mu.RLock()
	defer h.mu.RUnlock()
	keys := make([]SubscriptionKey, 0, len(h.subscriptions))
	for k := range h.subscriptions {
		keys = append(keys, k)
	}
	return keys
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Infof("[WS] hub stopped by ctx done.")
			return
		case client := <-h.register:
			h.clients[client] = true
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Infof("[WS] register client %v", client.ID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)

				// clean subscription
				for key := range client.Subscriptions {
					if clients, ok := h.subscriptions[key]; ok {
						delete(clients, client)
						if len(clients) == 0 {
							delete(h.subscriptions, key)
						}
					}
				}
				client.Conn.Close()
			}

			h.mu.Unlock()
			log.Infof("[WS] unregister client %v", client.ID)

		case message := <-h.broadcast:
			h.mu.RUnlock()
			for client := range h.clients {
				client.Send <- message
			}
			h.mu.RLock()
		}
	}
}

func (h *Hub) Subscribe(client *Client, key SubscriptionKey) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// create key in hub.subscriptions
	if h.subscriptions[key] == nil {
		h.subscriptions[key] = make(map[*Client]bool)
	}
	h.subscriptions[key][client] = true

	client.mu.Lock()
	client.Subscriptions[key] = true
	client.mu.Unlock()
	log.Infof("[WS] subscribe client %v, key:%v", client.ID, key)
}

func (h *Hub) Unsubscribe(client *Client, subKey SubscriptionKey) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.subscriptions[subKey]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.subscriptions, subKey)
		}
	}

	client.mu.Lock()
	delete(client.Subscriptions, subKey)
	client.mu.Unlock()

	log.Infof("[WS] unsubscribe client %v, key:%v", client.ID, subKey)
}

// BroadcastToSubscribers broadcast to all sub users.
func (h *Hub) BroadcastToSubscribers(key SubscriptionKey, data interface{}) {
	resp := WSResp{
		Channel:   key.Channel,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	message, err := json.Marshal(resp)
	if err != nil {
		log.Errorf("[WS] hub broadcast to all sub users failed.: %v", err)
		return
	}

	h.mu.RLock()
	log.Debugf("[BroadcastToSubscribers] inputKey:%v", key)
	clients, ok := h.subscriptions[key]
	h.mu.RUnlock()

	if !ok {
		return
	}

	for client := range clients {
		select {
		case client.Send <- message:
		default:
			h.unregister <- client
		}
	}
}
