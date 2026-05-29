package dextrades

import (
	"log"
	"strings"
	"time"

	"github.com/gofiber/contrib/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = 30 * time.Second
)

type Client struct {
	conn    *websocket.Conn
	send    chan []byte
	hub     *Hub
	address string
	chainID uint64
}

func NewClient(conn *websocket.Conn, hub *Hub, address string, chainID uint64) *Client {
	return &Client{
		conn:    conn,
		send:    make(chan []byte, 256),
		hub:     hub,
		address: strings.ToLower(strings.TrimSpace(address)),
		chainID: chainID,
	}
}

func (c *Client) Run() {
	c.hub.Register(c)
	defer func() {
		c.hub.Unregister(c)
		if err := c.conn.Close(); err != nil {
			log.Printf("dex trades ws close failed: %v", err)
		}
	}()

	done := make(chan struct{})
	go c.writePump(done)
	c.readPump()
	close(done)
}

func (c *Client) readPump() {
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (c *Client) writePump(done <-chan struct{}) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
