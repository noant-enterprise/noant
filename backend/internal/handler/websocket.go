package handler

import (
	"net/http"
	"sync"
	"time"

	"noant/internal/infrastructure"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WebSocketHub struct {
	clients       map[string]*websocket.Conn
	broadcast     chan WebSocketMessage
	register      chan *websocket.Conn
	unregister    chan *websocket.Conn
	mutex         sync.RWMutex
	logger        *infrastructure.Logger
	allowedOrigins []string
}

type WebSocketMessage struct {
	ConversationID string      `json:"conversation_id"`
	Type           string      `json:"type"`
	Data           interface{} `json:"data"`
}

func NewWebSocketHub(logger *infrastructure.Logger, allowedOrigins []string) *WebSocketHub {
	return &WebSocketHub{
		clients:        make(map[string]*websocket.Conn),
		broadcast:      make(chan WebSocketMessage, 256),
		register:       make(chan *websocket.Conn, 10),
		unregister:     make(chan *websocket.Conn, 10),
		logger:         logger,
		allowedOrigins: allowedOrigins,
	}
}

func (h *WebSocketHub) isOriginAllowed(origin string) bool {
	if origin == "" {
		return true // Non-browser clients (mobile, IoT)
	}
	for _, allowed := range h.allowedOrigins {
		if origin == allowed {
			return true
		}
	}
	return false
}

func (h *WebSocketHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client.RemoteAddr().String()] = client
			h.mutex.Unlock()
			h.logger.Info("WebSocket client connected", "addr", client.RemoteAddr().String())

		case client := <-h.unregister:
			h.mutex.Lock()
			delete(h.clients, client.RemoteAddr().String())
			h.mutex.Unlock()
			client.Close()
			h.logger.Info("WebSocket client disconnected", "addr", client.RemoteAddr().String())

		case msg := <-h.broadcast:
			h.mutex.RLock()
			for addr, client := range h.clients {
				if err := client.WriteJSON(msg); err != nil {
					h.logger.Warn("Failed to send WebSocket message", "addr", addr, "error", err)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

func (h *WebSocketHub) HandleWebSocket(c *gin.Context) {
	origin := c.Request.Header.Get("Origin")
	if !h.isOriginAllowed(origin) {
		h.logger.Warn("WebSocket origin not allowed", "origin", origin, "allowed", h.allowedOrigins)
		c.JSON(http.StatusForbidden, gin.H{"error": "origin not allowed"})
		return
	}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true }, // Already validated above
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("WebSocket upgrade failed", "error", err)
		return
	}

	h.register <- conn

	go func() {
		defer func() {
			if r := recover(); r != nil {
				h.logger.Error("WebSocket read panic recovered", "panic", r)
			}
			h.unregister <- conn
		}()

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		go func() {
			defer func() {
				if r := recover(); r != nil {
					h.logger.Error("WebSocket ping panic recovered", "panic", r)
				}
			}()
			for range ticker.C {
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}()

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					h.logger.Warn("WebSocket error", "error", err)
				}
				break
			}
		}
	}()
}

func (h *WebSocketHub) BroadcastMessage(msg WebSocketMessage) {
    h.logger.Info("BroadcastMessage called", "type", msg.Type, "conversation", msg.ConversationID, "clients", len(h.clients))
	h.logger.Info("BroadcastMessage called", "type", msg.Type, "conversation", msg.ConversationID, "clients", len(h.clients))
	select {
	case h.broadcast <- msg:
	default:
		h.logger.Warn("WebSocket broadcast channel full, dropping message")
	}
}