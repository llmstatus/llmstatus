package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	err := handleSubscription(client, msg)
	require.NoError(t, err)
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
	err := handleSubscription(client, subMsg)
	require.NoError(t, err)
	assert.True(t, client.subscriptions["provider:openai"])

	// Unsubscribe
	unsubMsg := SubscriptionMessage{
		Type:  "unsubscribe",
		Topic: "provider:openai",
	}
	err = handleSubscription(client, unsubMsg)
	require.NoError(t, err)
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
	err := handleSubscription(client, msg)
	require.NoError(t, err)

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
		err := handleSubscription(client, msg)
		require.NoError(t, err)
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

	// Spec requires ignoring invalid types and returning nil
	err := handleSubscription(client, msg)
	assert.NoError(t, err)
	// Invalid type should be ignored, no subscription created
	assert.False(t, client.subscriptions["provider:openai"])
}
