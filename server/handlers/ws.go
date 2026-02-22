package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // CORS handled by middleware
	},
}

// WSMessage represents a WebSocket message envelope.
type WSMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// SubscribePayload is sent by clients to subscribe to server channels.
type SubscribePayload struct {
	ServerIDs []string `json:"serverIds"`
	Channel   string   `json:"channel"`
}

// WSClient represents a connected WebSocket client.
type WSClient struct {
	conn      *websocket.Conn
	send      chan []byte
	hub       *WSHub
	serverIDs map[string]bool
	mu        sync.Mutex
}

// WSHub manages all connected WebSocket clients.
type WSHub struct {
	clients    map[*WSClient]bool
	register   chan *WSClient
	unregister chan *WSClient
	broadcast  chan []byte
	mu         sync.RWMutex
}

// Hub is the global WebSocket hub instance.
var Hub = &WSHub{
	clients:    make(map[*WSClient]bool),
	register:   make(chan *WSClient),
	unregister: make(chan *WSClient),
	broadcast:  make(chan []byte, 256),
}

func init() {
	go Hub.run()
}

func (h *WSHub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					go func(c *WSClient) { h.unregister <- c }(client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastMetrics sends a metrics update to clients subscribed to serverID.
func (h *WSHub) BroadcastMetrics(serverID string, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	msg := WSMessage{Type: "metrics", Payload: json.RawMessage(data)}
	raw, _ := json.Marshal(msg)

	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		client.mu.Lock()
		subscribed := client.serverIDs[serverID] || client.serverIDs["*"]
		client.mu.Unlock()
		if subscribed {
			select {
			case client.send <- raw:
			default:
			}
		}
	}
}

// BroadcastAlert sends an alert event to all connected clients.
func (h *WSHub) BroadcastAlert(payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	msg := WSMessage{Type: "alert", Payload: json.RawMessage(data)}
	raw, _ := json.Marshal(msg)
	h.broadcast <- raw
}

// BroadcastStatusChange sends a server status change to all clients.
func (h *WSHub) BroadcastStatusChange(serverID, status string) {
	payload := map[string]string{"serverId": serverID, "status": status}
	data, _ := json.Marshal(payload)
	msg := WSMessage{Type: "status_change", Payload: json.RawMessage(data)}
	raw, _ := json.Marshal(msg)
	h.broadcast <- raw
}

// HandleWS upgrades the HTTP connection to WebSocket.
func HandleWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &WSClient{
		conn:      conn,
		send:      make(chan []byte, 256),
		hub:       Hub,
		serverIDs: make(map[string]bool),
	}
	Hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *WSClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second)) //nolint:errcheck
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second)) //nolint:errcheck
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var msg WSMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "subscribe":
			var sub SubscribePayload
			if err := json.Unmarshal(msg.Payload, &sub); err != nil {
				continue
			}
			c.mu.Lock()
			for _, id := range sub.ServerIDs {
				c.serverIDs[id] = true
			}
			c.mu.Unlock()

		case "unsubscribe":
			var sub SubscribePayload
			if err := json.Unmarshal(msg.Payload, &sub); err != nil {
				continue
			}
			c.mu.Lock()
			for _, id := range sub.ServerIDs {
				delete(c.serverIDs, id)
			}
			c.mu.Unlock()

		case "ping":
			pong, _ := json.Marshal(WSMessage{Type: "pong", Payload: json.RawMessage(`{}`)})
			select {
			case c.send <- pong:
			default:
			}
		}
	}
}

func (c *WSClient) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second)) //nolint:errcheck
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{}) //nolint:errcheck
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second)) //nolint:errcheck
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
