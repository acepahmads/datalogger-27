package websocket

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"datalogger/internal/logger"
	"datalogger/pkg/sysinfo"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow local network origins
	},
}

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

type WSMessage struct {
	Type      string      `json:"type"` // "HEALTH_UPDATE", "ACTIVITY_LOG", "TASK_UPDATE", "ALARM"
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan WSMessage
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

var GlobalHub *Hub

func NewHub() *Hub {
	h := &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan WSMessage, 100),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
	GlobalHub = h
	return h
}

func (h *Hub) Run() {
	// Periodic telemetry ticker: 2-second interval, only runs if clients are connected
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			logger.Debug("WebSocket client connected. Total clients: %d", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			logger.Debug("WebSocket client disconnected. Total clients: %d", len(h.clients))

		case message := <-h.broadcast:
			data, err := json.Marshal(message)
			if err != nil {
				continue
			}
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- data:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()

		case <-ticker.C:
			h.mu.RLock()
			clientCount := len(h.clients)
			h.mu.RUnlock()

			if clientCount > 0 {
				info := sysinfo.GetInfo()
				msg := WSMessage{
					Type:      "HEALTH_UPDATE",
					Timestamp: time.Now(),
					Data:      info,
				}
				data, err := json.Marshal(msg)
				if err != nil {
					continue
				}
				h.mu.RLock()
				for client := range h.clients {
					select {
					case client.send <- data:
					default:
					}
				}
				h.mu.RUnlock()
			}
		}
	}
}

func (h *Hub) Broadcast(msg WSMessage) {
	select {
	case h.broadcast <- msg:
	default:
		// Queue full, drop non-critical WS messages to preserve edge CPU
	}
}

// ServeWS handles incoming websocket upgrade requests
func ServeWS(hub *Hub, c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("Failed to upgrade websocket: %v", err)
		return
	}

	client := &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 64),
	}

	hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(4096)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Flush queued messages to avoid separate network frames
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
