package adapters

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/llmstatus/llmstatus/internal/probes"
)

const (
	anthropicStreamingProbeType = "streaming"
	anthropicStreamingMaxTokens = 5
)

type anthropicStreamDelta struct {
	Type  string `json:"type"`
	Delta struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"delta"`
}

// ProbeStreaming issues a streaming messages request and records the time to
// first content_block_delta token (TTFT). DurationMs is TTFT, not total
// response time. A stream that ends without a text_delta is classified as
// ErrorClassMalformedBody.
func (p *anthropicProvider) ProbeStreaming(ctx context.Context, model string) (probes.ProbeResult, error) {
	started := time.Now()
	r := probes.ProbeResult{
		ProviderID: anthropicProviderID,
		Model:      model,
		ProbeType:  anthropicStreamingProbeType,
		StartedAt:  started.UTC(),
		RegionID:   p.region,
	}

	type streamRequest struct {
		Model     string             `json:"model"`
		MaxTokens int                `json:"max_tokens"`
		Messages  []anthropicMessage `json:"messages"`
		Stream    bool               `json:"stream"`
	}
	body, err := json.Marshal(streamRequest{
		Model:     model,
		MaxTokens: anthropicStreamingMaxTokens,
		Messages:  []anthropicMessage{{Role: "user", Content: "ping"}},
		Stream:    true,
	})
	if err != nil {
		r.DurationMs = time.Since(started).Milliseconds()
		r.ErrorClass = probes.ErrorClassUnknown
		r.ErrorDetail = truncate(err.Error(), anthropicErrorDetailMax)
		return r, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		r.DurationMs = time.Since(started).Milliseconds()
		r.ErrorClass = probes.ErrorClassUnknown
		r.ErrorDetail = truncate(err.Error(), anthropicErrorDetailMax)
		return r, err
	}
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", anthropicVersion)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", probeUserAgent)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := p.client.Do(req)
	if err != nil {
		r.DurationMs = time.Since(started).Milliseconds()
		r.ErrorClass = classifyNetError(err)
		r.ErrorDetail = truncate(err.Error(), anthropicErrorDetailMax)
		return r, nil
	}
	defer func() { _ = resp.Body.Close() }()

	r.HTTPStatus = resp.StatusCode
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		r.DurationMs = time.Since(started).Milliseconds()
		r.ErrorClass = classifyAnthropicStatus(resp.StatusCode)
		return r, nil
	}

	durationMs, gotToken := scanAnthropicStream(bufio.NewScanner(resp.Body), started)
	r.DurationMs = durationMs
	if !gotToken {
		r.ErrorClass = probes.ErrorClassMalformedBody
		r.ErrorDetail = "stream ended without content token"
		return r, nil
	}

	r.Success = true
	return r, nil
}

// scanAnthropicStream reads SSE lines until a text_delta token is found.
// Returns the elapsed ms and whether a token was received.
func scanAnthropicStream(scanner *bufio.Scanner, started time.Time) (int64, bool) {
	var currentEvent string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "event: ") {
			currentEvent = strings.TrimPrefix(line, "event: ")
			continue
		}
		if currentEvent != "content_block_delta" || !strings.HasPrefix(line, "data: ") {
			continue
		}
		var delta anthropicStreamDelta
		if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &delta); err != nil {
			continue
		}
		if delta.Delta.Type == "text_delta" && delta.Delta.Text != "" {
			return time.Since(started).Milliseconds(), true
		}
	}
	return time.Since(started).Milliseconds(), false
}
