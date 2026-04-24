package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProviderSubscription(t *testing.T) {
	hub := NewHub()
	client := &Client{
		hub:           hub,
		subscriptions: make(map[string]bool),
	}

	msg := SubscriptionMessage{
		Type:  "subscribe",
		Topic: "provider:openai",
	}

	handleSubscription(client, msg)
	assert.True(t, client.subscriptions["provider:openai"])
}

func TestProviderUnsubscription(t *testing.T) {
	hub := NewHub()
	client := &Client{
		hub:           hub,
		subscriptions: make(map[string]bool),
	}

	// Subscribe first
	subMsg := SubscriptionMessage{
		Type:  "subscribe",
		Topic: "provider:openai",
	}
	handleSubscription(client, subMsg)
	assert.True(t, client.subscriptions["provider:openai"])

	// Unsubscribe
	unsubMsg := SubscriptionMessage{
		Type:  "unsubscribe",
		Topic: "provider:openai",
	}
	handleSubscription(client, unsubMsg)
	assert.False(t, client.subscriptions["provider:openai"])
}

func TestBroadcastToTopic(t *testing.T) {
	hub := NewHub()

	client := &Client{
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	// Manually add client to hub without running the hub
	hub.mu.Lock()
	hub.clients[client] = true
	hub.mu.Unlock()

	// Subscribe to topic
	msg := SubscriptionMessage{
		Type:  "subscribe",
		Topic: "provider:openai",
	}
	handleSubscription(client, msg)

	// Broadcast to topic
	update := ProviderStatusUpdate{
		ProviderID: "openai",
		Status:     "operational",
		Latency:    150,
	}
	hub.BroadcastToTopic("provider:openai", update)

	// Verify message received
	select {
	case received := <-client.send:
		assert.NotNil(t, received)
	default:
		t.Fatal("expected message on send channel")
	}
}

func TestMultipleSubscriptions(t *testing.T) {
	hub := NewHub()
	client := &Client{
		hub:           hub,
		subscriptions: make(map[string]bool),
	}

	topics := []string{"provider:openai", "provider:anthropic", "incidents:all"}
	for _, topic := range topics {
		msg := SubscriptionMessage{
			Type:  "subscribe",
			Topic: topic,
		}
		handleSubscription(client, msg)
	}

	for _, topic := range topics {
		assert.True(t, client.subscriptions[topic], "expected subscription to %s", topic)
	}
}

func TestInvalidSubscriptionType(t *testing.T) {
	hub := NewHub()
	client := &Client{
		hub:           hub,
		subscriptions: make(map[string]bool),
	}

	msg := SubscriptionMessage{
		Type:  "invalid",
		Topic: "provider:openai",
	}

	// Spec requires ignoring invalid types — no subscription created
	handleSubscription(client, msg)
	assert.False(t, client.subscriptions["provider:openai"])
}
