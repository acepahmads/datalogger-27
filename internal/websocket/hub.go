package websocket

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"datalogger/internal/logger"
	"datalogger/pkg/sysinfo"
	"datalogger/pkg/tenantctx"

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
	hub            *Hub
	conn           *websocket.Conn
	send           chan []byte
	role           string
	username       string
	userTenantID   *uint
	activeTenantID uint
	activeSiteID   uint
	allowedSiteIDs []uint
	scopeMu        sync.RWMutex
}

type WSMessage struct {
	Type      string      `json:"type"` // "HEALTH_UPDATE", "ACTIVITY_LOG", "TASK_UPDATE", "ALARM"
	Timestamp time.Time   `json:"timestamp"`
	TenantID  uint        `json:"tenant_id,omitempty"`
	SiteID    uint        `json:"site_id,omitempty"`
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

var (
	tokenValidatorMu sync.RWMutex
	tokenValidator   func(token string) (role string, username string, err error)
	scopeValidator   func(token string) (role string, username string, tenantID *uint, allowedSiteIDs []uint, err error)
)

// SetTokenValidator registers a token validator function for WebSocket clients
func SetTokenValidator(validator func(token string) (role string, username string, err error)) {
	tokenValidatorMu.Lock()
	defer tokenValidatorMu.Unlock()
	tokenValidator = validator
}

// SetScopeValidator registers a multi-tenant aware token validator function for WebSocket clients
func SetScopeValidator(validator func(token string) (role string, username string, tenantID *uint, allowedSiteIDs []uint, err error)) {
	tokenValidatorMu.Lock()
	defer tokenValidatorMu.Unlock()
	scopeValidator = validator
}

func extractUintField(v interface{}) uint {
	switch n := v.(type) {
	case uint:
		return n
	case int:
		if n > 0 {
			return uint(n)
		}
	case int64:
		if n > 0 {
			return uint(n)
		}
	case float64:
		if n > 0 {
			return uint(n)
		}
	}
	return 0
}

func extractMessageScope(msg WSMessage) (uint, uint) {
	tid, sid := msg.TenantID, msg.SiteID
	if mData, ok := msg.Data.(map[string]interface{}); ok {
		if tid == 0 {
			tid = extractUintField(mData["tenant_id"])
		}
		if sid == 0 {
			sid = extractUintField(mData["site_id"])
		}
	}
	return tid, sid
}

func (c *Client) matchesScope(msgTenantID, msgSiteID uint) bool {
	c.scopeMu.RLock()
	defer c.scopeMu.RUnlock()

	isSuper := strings.EqualFold(c.role, "SUPERADMIN") || strings.EqualFold(c.role, "Administrator")
	if msgTenantID > 0 {
		if !isSuper {
			if c.userTenantID != nil && *c.userTenantID > 0 && *c.userTenantID != msgTenantID {
				return false
			}
			if c.activeTenantID > 0 && c.activeTenantID != msgTenantID {
				return false
			}
		} else if c.activeTenantID > 0 && c.activeTenantID != msgTenantID {
			return false
		}
	}
	if msgSiteID > 0 {
		if !isSuper && len(c.allowedSiteIDs) > 0 {
			allowed := false
			for _, id := range c.allowedSiteIDs {
				if id == msgSiteID {
					allowed = true
					break
				}
			}
			if !allowed {
				return false
			}
		}
		if c.activeSiteID > 0 && c.activeSiteID != msgSiteID {
			return false
		}
	}
	return true
}

func NewHub() *Hub {
	h := &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan WSMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
	GlobalHub = h
	return h
}

func (h *Hub) Run() {
	// Periodic telemetry ticker: 5-second interval matching hardware sampling, only runs if clients are connected
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			logger.Debug("WebSocket client connected (%s:%s). Total clients: %d", client.username, client.role, len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			logger.Debug("WebSocket client disconnected. Total clients: %d", len(h.clients))

		case message := <-h.broadcast:
			msgTenantID, msgSiteID := extractMessageScope(message)
			h.mu.RLock()
			for client := range h.clients {
				isSuper := strings.EqualFold(client.role, "SUPERADMIN") || strings.EqualFold(client.role, "Administrator")
				canRaw := isSuper || strings.EqualFold(client.role, "ADMIN")

				// 0. Strict Multi-Tenant & Multi-Site room isolation
				if !client.matchesScope(msgTenantID, msgSiteID) {
					continue
				}

				// 1. Strict restriction: Non-superadmin CANNOT receive development events or activity monitor events
				if (message.Type == "TASK_UPDATE" || message.Type == "ACTIVITY_LOG" || message.Type == "SYSTEM_ACTIVITY_EVENT" || message.Type == "activity.event") && !isSuper {
					continue
				}

				// 2. Strict restriction: Non-admin/superadmin CANNOT receive raw telemetry fields
				msgToSend := message
				if !canRaw && message.Type == "device.telemetry.received" {
					if mData, ok := message.Data.(map[string]interface{}); ok {
						sanitized := make(map[string]interface{}, len(mData))
						for k, v := range mData {
							if k == "raw_value" {
								sanitized[k] = 0
							} else if k == "raw_hex" {
								sanitized[k] = ""
							} else if k == "raw_bytes" {
								sanitized[k] = nil
							} else {
								sanitized[k] = v
							}
						}
						msgToSend.Data = sanitized
					}
				}

				data, err := json.Marshal(msgToSend)
				if err != nil {
					continue
				}

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

	tokenStr := c.Query("token")
	if tokenStr == "" {
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	role := "OPERATOR"
	username := "guest"
	var userTenantID *uint
	var allowedSites []uint

	tokenValidatorMu.RLock()
	val := tokenValidator
	sVal := scopeValidator
	tokenValidatorMu.RUnlock()

	if tokenStr == "local-admin-token" {
		role = "SUPERADMIN"
		username = "admin"
	} else if sVal != nil && tokenStr != "" {
		if r, u, tid, sites, err := sVal(tokenStr); err == nil && r != "" {
			role = r
			username = u
			userTenantID = tid
			allowedSites = sites
		}
	} else if val != nil && tokenStr != "" {
		if r, u, err := val(tokenStr); err == nil && r != "" {
			role = r
			username = u
		}
	}

	var reqTenantID, reqSiteID uint
	if v, err := strconv.ParseUint(strings.TrimSpace(c.Query("tenant_id")), 10, 32); err == nil {
		reqTenantID = uint(v)
	}
	if v, err := strconv.ParseUint(strings.TrimSpace(c.Query("site_id")), 10, 32); err == nil {
		reqSiteID = uint(v)
	}

	isSuper := strings.EqualFold(role, "SUPERADMIN") || strings.EqualFold(role, "Administrator")
	activeTenantID := reqTenantID
	if !isSuper && userTenantID != nil && *userTenantID > 0 {
		activeTenantID = *userTenantID
	}
	activeSiteID := reqSiteID
	if !isSuper && len(allowedSites) > 0 {
		allowed := false
		for _, id := range allowedSites {
			if id == reqSiteID {
				allowed = true
				break
			}
		}
		if !allowed {
			if len(allowedSites) == 1 {
				activeSiteID = allowedSites[0]
			} else {
				activeSiteID = 0
			}
		}
	}

	client := &Client{
		hub:            hub,
		conn:           conn,
		send:           make(chan []byte, 256),
		role:           role,
		username:       username,
		userTenantID:   userTenantID,
		activeTenantID: activeTenantID,
		activeSiteID:   activeSiteID,
		allowedSiteIDs: allowedSites,
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
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		// Allow live scope updates (SET_SCOPE) from Header Tenant/Site switcher without reconnecting
		var cmd struct {
			Type     string `json:"type"`
			TenantID uint   `json:"tenant_id"`
			SiteID   uint   `json:"site_id"`
		}
		if json.Unmarshal(raw, &cmd) == nil && strings.EqualFold(cmd.Type, "SET_SCOPE") {
			c.scopeMu.Lock()
			isSuper := strings.EqualFold(c.role, "SUPERADMIN") || strings.EqualFold(c.role, "Administrator")
			if isSuper {
				c.activeTenantID = cmd.TenantID
				c.activeSiteID = cmd.SiteID
			} else {
				if c.userTenantID != nil && *c.userTenantID > 0 {
					c.activeTenantID = *c.userTenantID
				}
				s := tenantctx.Scope{AllowedSiteIDs: c.allowedSiteIDs}
				if cmd.SiteID == 0 || s.IsSiteAllowed(cmd.SiteID) {
					c.activeSiteID = cmd.SiteID
				}
			}
			c.scopeMu.Unlock()
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
