package probes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPSink implements ResultSink by POSTing each ProbeResult as JSON to a
// remote ingest URL. It is used by cmd/prober when INGEST_URL is configured.
type HTTPSink struct {
	// URL is the ingest endpoint, e.g. "http://ingest:8080/v1/probe".
	URL    string
	Client *http.Client
}

// NewHTTPSink returns an HTTPSink with a sensible default HTTP client.
func NewHTTPSink(url string) *HTTPSink {
	return &HTTPSink{
		URL:    url,
		Client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Receive encodes r as JSON and POSTs it to the ingest URL.
// Non-2xx responses are treated as errors.
func (s *HTTPSink) Receive(ctx context.Context, r ProbeResult) error {
	body, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("httpsink: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("httpsink: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("httpsink: post: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return fmt.Errorf("httpsink: ingest returned %d: %s", resp.StatusCode, bytes.TrimSpace(detail))
	}
	return nil
}
