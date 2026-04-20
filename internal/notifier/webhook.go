package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var webhookClient = &http.Client{Timeout: 10 * time.Second}

// deliverWebhook POSTs payload to url with 3 attempts and exponential backoff (1s, 4s, 16s).
func deliverWebhook(ctx context.Context, url string, payload webhookPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook: marshal: %w", err)
	}

	delays := []time.Duration{0, time.Second, 4 * time.Second, 16 * time.Second}
	var lastErr error
	for attempt, delay := range delays {
		if delay > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
		lastErr = postWebhook(ctx, url, body)
		if lastErr == nil {
			return nil
		}
		if attempt < len(delays)-1 {
			// log transient failure; will retry
			_ = lastErr
		}
	}
	return fmt.Errorf("webhook: all attempts failed: %w", lastErr)
}

func postWebhook(ctx context.Context, url string, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "llmstatus-notifier/1")

	resp, err := webhookClient.Do(req)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("remote returned %d", resp.StatusCode)
	}
	return nil
}
