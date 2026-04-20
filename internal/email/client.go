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

type Client struct {
	apiKey string
	from   string
	http   *http.Client
}

func New(apiKey, from string) *Client {
	return &Client{
		apiKey: apiKey,
		from:   from,
		http:   &http.Client{Timeout: 10 * time.Second},
	}
}

type Message struct {
	To      string
	Subject string
	HTML    string
	Text    string
}

func (c *Client) Send(ctx context.Context, msg Message) error {
	payload := map[string]any{
		"from":    c.from,
		"to":      []string{msg.To},
		"subject": msg.Subject,
		"html":    msg.HTML,
		"text":    msg.Text,
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendAPI, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("email: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("email: send: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("email: resend returned %d", resp.StatusCode)
	}
	return nil
}
