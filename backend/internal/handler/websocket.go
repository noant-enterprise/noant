package handler

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"noant/internal/infrastructure"

	sentry "github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type wsClient struct {
	conn   *websocket.Conn
	mu     sync.Mutex
	send   chan []byte
	closeOnce sync.Once
}

func (c *wsClient) writeMessage(messageType int, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteMessage(messageType, data)
}

func (c *wsClient) writeLoop() {
	for data := range c.send {
		c.mu.Lock()
		err := c.conn.WriteMessage(websocket.TextMessage, data)
		c.mu.Unlock()
		if err != nil {
			return
		}
	}
}

func (c *wsClient) close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.Close()
}

type WebSocketHub struct {
	clients        map[string]*wsClient
	adminClients   map[string]*wsClient
	broadcast      chan WebSocketMessage
	adminBroadcast chan AdminEvent
	register       chan *wsClient
	unregister     chan *wsClient
	adminRegister  chan *wsClient
	adminUnreg     chan *wsClient
	mutex          sync.RWMutex
	logger         *infrastructure.Logger
	allowedOrigins []string
}

type WebSocketMessage struct {
	ConversationID string      `json:"conversation_id"`
	Type           string      `json:"type"`
	Data           interface{} `json:"data"`
}

type AdminEvent struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

func NewWebSocketHub(logger *infrastructure.Logger, allowedOrigins []string) *WebSocketHub {
	return &WebSocketHub{
		clients:        make(map[string]*wsClient),
		adminClients:   make(map[string]*wsClient),
		broadcast:      make(chan WebSocketMessage, 256),
		adminBroadcast: make(chan AdminEvent, 256),
		register:       make(chan *wsClient, 10),
		unregister:     make(chan *wsClient, 10),
		adminRegister:  make(chan *wsClient, 10),
		adminUnreg:     make(chan *wsClient, 10),
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
	defer func() {
		if r := recover(); r != nil {
			h.logger.Error("Panic in WebSocket hub", "recover", r)
			sentry.CaptureMessage("WebSocket hub panic recovered")
			go h.Run()
		}
	}()
	for {
		select {
		case client := <-h.register:
			go client.writeLoop()
			h.mutex.Lock()
			h.clients[client.conn.RemoteAddr().String()] = client
			h.mutex.Unlock()
			h.logger.Info("WebSocket client connected", "addr", client.conn.RemoteAddr().String())

		case client := <-h.unregister:
			h.mutex.Lock()
			delete(h.clients, client.conn.RemoteAddr().String())
			h.mutex.Unlock()
			client.closeOnce.Do(func() { close(client.send); _ = client.close() })
			h.logger.Info("WebSocket client disconnected", "addr", client.conn.RemoteAddr().String())

		case client := <-h.adminRegister:
			go client.writeLoop()
			h.mutex.Lock()
			h.adminClients[client.conn.RemoteAddr().String()] = client
			h.mutex.Unlock()
			h.logger.Info("Admin WebSocket connected", "addr", client.conn.RemoteAddr().String())

		case client := <-h.adminUnreg:
			h.mutex.Lock()
			delete(h.adminClients, client.conn.RemoteAddr().String())
			h.mutex.Unlock()
			client.closeOnce.Do(func() { close(client.send); _ = client.close() })
			h.logger.Info("Admin WebSocket disconnected", "addr", client.conn.RemoteAddr().String())

		case msg := <-h.broadcast:
			data, _ := json.Marshal(msg)
			h.mutex.RLock()
			for _, client := range h.clients {
				select {
				case client.send <- data:
				default:
				}
			}
			h.mutex.RUnlock()

		case event := <-h.adminBroadcast:
			data, _ := json.Marshal(event)
			h.mutex.RLock()
			for _, client := range h.adminClients {
				select {
				case client.send <- data:
				default:
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

	conn.SetReadLimit(65536) // 64KB max message size

	client := &wsClient{conn: conn, send: make(chan []byte, 64)}
	h.register <- client

	go func() {
		defer func() {
			if r := recover(); r != nil {
				h.logger.Error("WebSocket read panic recovered", "panic", r)
			}
			h.unregister <- client
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
				if err := client.writeMessage(websocket.PingMessage, nil); err != nil {
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
	select {
	case h.broadcast <- msg:
	default:
		h.logger.Warn("WebSocket broadcast channel full, dropping message")
	}
}

func (h *WebSocketHub) HandleAdminWebSocket(c *gin.Context) {
	origin := c.Request.Header.Get("Origin")
	if !h.isOriginAllowed(origin) {
		c.JSON(http.StatusForbidden, gin.H{"error": "origin not allowed"})
		return
	}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Admin WebSocket upgrade failed", "error", err)
		return
	}

	conn.SetReadLimit(65536)

	client := &wsClient{conn: conn, send: make(chan []byte, 64)}
	h.adminRegister <- client

	go func() {
		defer func() {
			if r := recover(); r != nil {
				h.logger.Error("Admin WS read panic", "panic", r)
			}
			h.adminUnreg <- client
		}()

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		go func() {
			defer func() {
				if r := recover(); r != nil {
					h.logger.Error("Admin WS ping panic", "panic", r)
				}
			}()
			for range ticker.C {
				if err := client.writeMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}()

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "admin websocket connected"})
}

func (h *WebSocketHub) BroadcastAdminEvent(eventType string, payload interface{}) {
	event := AdminEvent{Type: eventType, Payload: payload}
	data, _ := json.Marshal(event)
	h.mutex.RLock()
	for _, client := range h.adminClients {
		select {
		case client.send <- data:
		default:
		}
	}
	h.mutex.RUnlock()
}