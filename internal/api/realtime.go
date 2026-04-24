package api

import (
	"encoding/json"
	"log/slog"
	"time"
)

// SubscriptionMessage represents a subscription request from a client.
type SubscriptionMessage struct {
	Type  string `json:"type"`
	Topic string `json:"topic"`
}

// StatusUpdate represents a real-time status update (spec-required struct).
type StatusUpdate struct {
	Type       string `json:"type"`
	ProviderID string `json:"provider_id"`
	Status     string `json:"status"`
	Timestamp  int64  `json:"timestamp"`
}

// RealtimeMessage represents a message sent over WebSocket to clients.
type RealtimeMessage struct {
	Type      string          `json:"type"`
	Channel   string          `json:"channel,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data,omitempty"`
}

// ProviderStatusUpdate represents a real-time provider status change.
type ProviderStatusUpdate struct {
	ProviderID string    `json:"provider_id"`
	Status     string    `json:"status"`
	Latency    int64     `json:"latency_ms"`
	Timestamp  time.Time `json:"timestamp"`
}

// IncidentUpdate represents a real-time incident change.
type IncidentUpdate struct {
	IncidentID string    `json:"incident_id"`
	ProviderID string    `json:"provider_id"`
	Status     string    `json:"status"`
	Severity   string    `json:"severity"`
	Timestamp  time.Time `json:"timestamp"`
}

// RealtimeManager handles real-time subscriptions and broadcasts.
type RealtimeManager struct {
	hub *Hub
}

// NewRealtimeManager creates a new RealtimeManager instance.
func NewRealtimeManager(hub *Hub) *RealtimeManager {
	return &RealtimeManager{
		hub: hub,
	}
}

// BroadcastProviderStatus sends a provider status update to all subscribed clients.
func (rm *RealtimeManager) BroadcastProviderStatus(update ProviderStatusUpdate) {
	msg := RealtimeMessage{
		Type:      "provider_status",
		Channel:   "provider:" + update.ProviderID,
		Timestamp: time.Now().UTC(),
	}

	data, err := json.Marshal(update)
	if err != nil {
		slog.Error("realtime: failed to marshal provider status", "err", err)
		return
	}
	msg.Data = data

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		slog.Error("realtime: failed to marshal message", "err", err)
		return
	}

	rm.hub.BroadcastToSubscribers(msg.Channel, msgBytes)
}

// BroadcastIncident sends an incident update to all subscribed clients.
func (rm *RealtimeManager) BroadcastIncident(update IncidentUpdate) {
	msg := RealtimeMessage{
		Type:      "incident",
		Channel:   "incidents:" + update.ProviderID,
		Timestamp: time.Now().UTC(),
	}

	data, err := json.Marshal(update)
	if err != nil {
		slog.Error("realtime: failed to marshal incident", "err", err)
		return
	}
	msg.Data = data

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		slog.Error("realtime: failed to marshal message", "err", err)
		return
	}

	rm.hub.BroadcastToSubscribers(msg.Channel, msgBytes)
	// Also broadcast to global incidents channel
	rm.hub.BroadcastToSubscribers("incidents", msgBytes)
}

// BroadcastGlobalStatus sends a global status update to all connected clients.
func (rm *RealtimeManager) BroadcastGlobalStatus(data interface{}) {
	msg := RealtimeMessage{
		Type:      "global_status",
		Channel:   "global",
		Timestamp: time.Now().UTC(),
	}

	msgData, err := json.Marshal(data)
	if err != nil {
		slog.Error("realtime: failed to marshal global status", "err", err)
		return
	}
	msg.Data = msgData

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		slog.Error("realtime: failed to marshal message", "err", err)
		return
	}

	rm.hub.Broadcast(msgBytes)
}

// ClientCount returns the number of connected clients.
func (rm *RealtimeManager) ClientCount() int {
	rm.hub.mu.RLock()
	defer rm.hub.mu.RUnlock()
	return len(rm.hub.clients)
}

// SubscriberCount returns the number of clients subscribed to a specific channel.
func (rm *RealtimeManager) SubscriberCount(channel string) int {
	rm.hub.mu.RLock()
	defer rm.hub.mu.RUnlock()

	count := 0
	for client := range rm.hub.clients {
		if client.IsSubscribed(channel) {
			count++
		}
	}
	return count
}

// handleSubscription processes subscription and unsubscription requests.
func handleSubscription(client *Client, msg SubscriptionMessage) error {
	switch msg.Type {
	case "subscribe":
		client.subscriptions[msg.Topic] = true
		slog.Info("client subscribed", "topic", msg.Topic)
	case "unsubscribe":
		delete(client.subscriptions, msg.Topic)
		slog.Info("client unsubscribed", "topic", msg.Topic)
	}
	return nil
}

// BroadcastToTopic sends a message to all clients subscribed to a specific topic.
func (h *Hub) BroadcastToTopic(topic string, data interface{}) {
	message, err := json.Marshal(data)
	if err != nil {
		slog.Error("subscription: failed to marshal broadcast data", "err", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		if client.IsSubscribed(topic) {
			select {
			case client.send <- message:
			default:
				close(client.send)
				delete(h.clients, client)
			}
		}
	}
}
