// Package email provides a simple HTTP-based email sending client.
package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const resendAPI = "https://api.resend.com/emails"

// Sender is the interface satisfied by *Client; useful for test mocking.
type Sender interface {
	Send(ctx context.Context, msg Message) error
}

// Client sends transactional emails via HTTP.
type Client struct {
	apiKey  string
	from    string
	baseURL string
	http    *http.Client
}

// New creates an email client from environment config.
func New(apiKey, from string) *Client {
	return &Client{
		apiKey:  apiKey,
		from:    from,
		baseURL: resendAPI,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

// NewWithBaseURL creates a Client pointing at a custom URL (for tests).
func NewWithBaseURL(baseURL, apiKey, from string) *Client {
	return &Client{
		apiKey:  apiKey,
		from:    from,
		baseURL: baseURL,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

// Message holds the fields for an outbound email.
type Message struct {
	To      string
	Subject string
	HTML    string
	Text    string
}

// Send delivers a single email message.
func (c *Client) Send(ctx context.Context, msg Message) error {
	payload := map[string]any{
		"from":    c.from,
		"to":      []string{msg.To},
		"subject": msg.Subject,
		"html":    msg.HTML,
		"text":    msg.Text,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("email: marshal payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("email: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("email: send: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode >= 300 {
		return fmt.Errorf("email: resend returned %d", resp.StatusCode)
	}
	return nil
}
