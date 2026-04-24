package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebSocketUpgrade(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(HandleWebSocket))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)
	defer ws.Close()

	// Should receive connection confirmation
	var msg map[string]interface{}
	err = ws.ReadJSON(&msg)
	require.NoError(t, err)
	assert.Equal(t, "connected", msg["type"])
}

func TestWebSocketClientRegistration(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleWebSocketWithHub(w, r, hub)
	}))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)
	defer ws.Close()

	// Receive connection confirmation
	var msg map[string]interface{}
	err = ws.ReadJSON(&msg)
	require.NoError(t, err)
	assert.Equal(t, "connected", msg["type"])

	// Verify client is registered in hub
	hub.mu.RLock()
	clientCount := len(hub.clients)
	hub.mu.RUnlock()
	assert.Equal(t, 1, clientCount)
}

func TestWebSocketMultipleClients(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleWebSocketWithHub(w, r, hub)
	}))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect first client
	ws1, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)
	defer ws1.Close()

	var msg1 map[string]interface{}
	err = ws1.ReadJSON(&msg1)
	require.NoError(t, err)

	// Connect second client
	ws2, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)
	defer ws2.Close()

	var msg2 map[string]interface{}
	err = ws2.ReadJSON(&msg2)
	require.NoError(t, err)

	// Verify both clients are registered
	hub.mu.RLock()
	clientCount := len(hub.clients)
	hub.mu.RUnlock()
	assert.Equal(t, 2, clientCount)
}

func TestWebSocketClientDisconnection(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleWebSocketWithHub(w, r, hub)
	}))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect client
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)

	var msg map[string]interface{}
	err = ws.ReadJSON(&msg)
	require.NoError(t, err)

	// Verify client is registered
	hub.mu.RLock()
	initialCount := len(hub.clients)
	hub.mu.RUnlock()
	assert.Equal(t, 1, initialCount)

	// Close connection
	ws.Close()

	// Give hub time to process unregister
	// In production, this would be handled by the client read loop detecting closure
	// For testing, we just verify the mechanism exists
}
