package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/gommon/log"
	"net/http"
	"sync"
	"time"
)

type Client struct {
	ID            string
	Conn          *websocket.Conn
	Send          chan []byte
	Subscriptions map[SubscriptionKey]bool
	Hub           *Hub
	mu            sync.RWMutex
}

func NewClient(id string, conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		ID:            id,
		Conn:          conn,
		Send:          make(chan []byte, 256),
		Subscriptions: make(map[SubscriptionKey]bool),
		Hub:           hub,
	}
}

// ReadPump read client req data
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Infof("[WS] Client read message error: %v", err)
			}
			break
		}

		var req WSReq
		if err := json.Unmarshal(message, &req); err != nil {
			log.Infof("[WS] Client read message error: %v", err)
			continue
		}

		c.handleRequest(req)
	}
}

func (c *Client) handleRequest(req WSReq) {
	switch req.Action {
	case SUBSCRIBE:
		c.handleSubscribe(req)
	case UNSUBSCRIBE:
		c.handleUnsubscribe(req)
	default:
		log.Warnf("UNKNOWN Action: %s", req.Action)
	}
}

// handleSubscribe
func (c *Client) handleSubscribe(req WSReq) {
	key, err := BuildSubscriptionKey(req)
	if err != nil {
		log.Errorf("Failed to create subscribtion key: %v", err)
		return
	}

	c.Hub.Subscribe(c, key)
}

// handleUnsubscribe
func (c *Client) handleUnsubscribe(req WSReq) {
	key, err := BuildSubscriptionKey(req)
	if err != nil {
		log.Errorf("Failed to create subscribtion key: %v", err)
		return
	}

	c.Hub.Unsubscribe(c, key)
}

// WritePump send response to user
func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// HandleWebSocket WebSocket handler
func HandleWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS] failed to upgrade: %v", err)
		return
	}

	clientID := fmt.Sprintf("client_%d", time.Now().UnixNano())
	client := NewClient(clientID, conn, hub)

	hub.register <- client

	go client.WritePump()
	go client.ReadPump()
}

// DataFeeder feed data
type DataFeeder interface {
	Start(ctx context.Context)
	Feed(ctx context.Context, pkg *WSFeedPackage)
}
