package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Validate origin to prevent CSRF attacks
		origin := r.Header.Get("Origin")
		if origin == "" {
			// Allow requests without Origin header (same-origin requests)
			return true
		}

		// Parse origin and request host
		originURL := strings.TrimPrefix(origin, "http://")
		originURL = strings.TrimPrefix(originURL, "https://")
		originHost := strings.Split(originURL, "/")[0]

		requestHost := r.Host
		if strings.Contains(requestHost, ":") {
			requestHost = strings.Split(requestHost, ":")[0]
		}
		if strings.Contains(originHost, ":") {
			originHost = strings.Split(originHost, ":")[0]
		}

		// Allow same-origin and localhost for development
		return originHost == requestHost || originHost == "localhost" || originHost == "127.0.0.1"
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// globalHub is the shared hub instance for all WebSocket connections
var globalHub *Hub
var hubOnce sync.Once

// Hub maintains the set of active client connections and broadcasts messages to them.
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

// Client represents a WebSocket connection to a client.
type Client struct {
	hub           *Hub
	conn          *websocket.Conn
	send          chan []byte
	subscriptions map[string]bool
	mu            sync.RWMutex
}

// NewHub creates a new Hub instance.
func NewHub() *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// GetGlobalHub returns the shared global hub instance, creating it if necessary.
func GetGlobalHub() *Hub {
	hubOnce.Do(func() {
		globalHub = NewHub()
		go globalHub.Run()
	})
	return globalHub
}

// Shutdown gracefully shuts down the hub and closes all client connections.
func (h *Hub) Shutdown(ctx context.Context) error {
	h.cancel()

	// Wait for all clients to disconnect or context to timeout
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		h.mu.RLock()
		clientCount := len(h.clients)
		h.mu.RUnlock()

		if clientCount == 0 {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

// Run starts the hub's event loop, processing client registrations, unregistrations,
// and broadcasts. This should be run in a goroutine. It exits when the context is cancelled.
func (h *Hub) Run() {
	for {
		select {
		case <-h.ctx.Done():
			// Graceful shutdown: close all client connections
			h.mu.Lock()
			for client := range h.clients {
				close(client.send)
				_ = client.conn.Close()
			}
			h.clients = make(map[*Client]bool)
			h.mu.Unlock()
			return

		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			slog.Debug("websocket: client registered", "total_clients", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			slog.Debug("websocket: client unregistered", "total_clients", len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Client's send channel is full, skip this message
					slog.Warn("websocket: client send channel full, dropping message")
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends a message to all connected clients.
func (h *Hub) Broadcast(message []byte) {
	select {
	case h.broadcast <- message:
	default:
		slog.Warn("websocket: broadcast channel full, dropping message")
	}
}

// HandleWebSocket upgrades an HTTP connection to WebSocket and manages the client.
// It uses the shared global hub instance for broadcasting across all connections.
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	hub := GetGlobalHub()
	HandleWebSocketWithHub(w, r, hub)
}

// HandleWebSocketWithHub upgrades an HTTP connection to WebSocket using the provided hub.
func HandleWebSocketWithHub(w http.ResponseWriter, r *http.Request, hub *Hub) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket: upgrade failed", "err", err)
		return
	}

	client := &Client{
		hub:           hub,
		conn:          conn,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	hub.register <- client

	// Send connection confirmation
	confirmMsg := map[string]string{"type": "connected"}
	if err := conn.WriteJSON(confirmMsg); err != nil {
		slog.Error("websocket: failed to send connection confirmation", "err", err)
		_ = conn.Close()
		hub.unregister <- client
		return
	}

	// Start goroutines for reading and writing
	go client.readPump()
	go client.writePump()
}

// readPump reads messages from the WebSocket connection and processes them.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		_ = c.conn.Close()
	}()

	_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	})

	for {
		var msg map[string]interface{}
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("websocket: read error", "err", err)
			}
			break
		}

		// Process message based on type
		if msgType, ok := msg["type"].(string); ok {
			switch msgType {
			case "subscribe":
				c.handleSubscribe(msg)
			case "unsubscribe":
				c.handleUnsubscribe(msg)
			case "ping":
				// Send pong response using proper JSON marshaling
				pongMsg := map[string]string{"type": "pong"}
				data, err := json.Marshal(pongMsg)
				if err != nil {
					slog.Error("websocket: failed to marshal pong message", "err", err)
					continue
				}
				c.send <- data
			default:
				slog.Debug("websocket: unknown message type", "type", msgType)
			}
		}
	}
}

// writePump writes messages from the client's send channel to the WebSocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleSubscribe processes subscription requests from clients.
func (c *Client) handleSubscribe(msg map[string]interface{}) {
	if channel, ok := msg["channel"].(string); ok {
		// Validate channel name: alphanumeric, underscore, colon, hyphen only
		if !isValidChannelName(channel) {
			slog.Warn("websocket: invalid channel name", "channel", channel)
			errMsg := map[string]interface{}{
				"type":  "error",
				"error": "invalid channel name",
			}
			data, _ := json.Marshal(errMsg)
			c.send <- data
			return
		}

		c.mu.Lock()
		c.subscriptions[channel] = true
		c.mu.Unlock()
		slog.Debug("websocket: client subscribed", "channel", channel)

		// Send subscription confirmation
		confirmMsg := map[string]interface{}{
			"type":    "subscribed",
			"channel": channel,
		}
		data, err := json.Marshal(confirmMsg)
		if err != nil {
			slog.Error("websocket: failed to marshal subscription confirmation", "err", err)
			return
		}
		c.send <- data
	}
}

// handleUnsubscribe processes unsubscription requests from clients.
func (c *Client) handleUnsubscribe(msg map[string]interface{}) {
	if channel, ok := msg["channel"].(string); ok {
		c.mu.Lock()
		delete(c.subscriptions, channel)
		c.mu.Unlock()
		slog.Debug("websocket: client unsubscribed", "channel", channel)

		// Send unsubscription confirmation
		confirmMsg := map[string]interface{}{
			"type":    "unsubscribed",
			"channel": channel,
		}
		data, err := json.Marshal(confirmMsg)
		if err != nil {
			slog.Error("websocket: failed to marshal unsubscription confirmation", "err", err)
			return
		}
		c.send <- data
	}
}

// isValidChannelName validates that a channel name contains only safe characters.
func isValidChannelName(channel string) bool {
	if len(channel) == 0 || len(channel) > 256 {
		return false
	}
	for _, r := range channel {
		if !isChannelRune(r) {
			return false
		}
	}
	return true
}

func isChannelRune(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '_' || r == ':' || r == '-'
}

// IsSubscribed checks if the client is subscribed to a channel.
func (c *Client) IsSubscribed(channel string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.subscriptions[channel]
}

// BroadcastToSubscribers sends a message to all clients subscribed to a specific channel.
func (h *Hub) BroadcastToSubscribers(channel string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		if client.IsSubscribed(channel) {
			select {
			case client.send <- message:
			default:
				slog.Warn("websocket: client send channel full, dropping message", "channel", channel)
			}
		}
	}
}
