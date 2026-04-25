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
	openaiStreamingProbeType = "streaming"
	openaiStreamingMaxTokens = 5
)

type openaiStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

// ProbeStreaming issues a streaming chat completion and records the time to
// first non-empty content token (TTFT). DurationMs is TTFT, not total response
// time. A stream that ends with [DONE] before producing any content token is
// classified as ErrorClassMalformedBody.
func (p *openaiProvider) ProbeStreaming(ctx context.Context, model string) (probes.ProbeResult, error) {
	if isOpenAIEmbeddingModel(model) {
		return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: openaiProviderID, ProbeType: "streaming"}
	}
	started := time.Now()
	r := probes.ProbeResult{
		ProviderID: openaiProviderID,
		Model:      model,
		ProbeType:  openaiStreamingProbeType,
		StartedAt:  started.UTC(),
		RegionID:   p.region,
	}

	type streamRequest struct {
		Model     string              `json:"model"`
		Messages  []openaiChatMessage `json:"messages"`
		MaxTokens int                 `json:"max_tokens"`
		Stream    bool                `json:"stream"`
	}
	body, err := json.Marshal(streamRequest{
		Model:     model,
		Messages:  []openaiChatMessage{{Role: "user", Content: "ping"}},
		MaxTokens: openaiStreamingMaxTokens,
		Stream:    true,
	})
	if err != nil {
		r.DurationMs = time.Since(started).Milliseconds()
		r.ErrorClass = probes.ErrorClassUnknown
		r.ErrorDetail = truncate(err.Error(), openaiErrorDetailMax)
		return r, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		r.DurationMs = time.Since(started).Milliseconds()
		r.ErrorClass = probes.ErrorClassUnknown
		r.ErrorDetail = truncate(err.Error(), openaiErrorDetailMax)
		return r, err
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", probeUserAgent)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := p.client.Do(req)
	if err != nil {
		r.DurationMs = time.Since(started).Milliseconds()
		r.ErrorClass = classifyNetError(err)
		r.ErrorDetail = truncate(err.Error(), openaiErrorDetailMax)
		return r, nil
	}
	defer func() { _ = resp.Body.Close() }()

	r.HTTPStatus = resp.StatusCode
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		r.DurationMs = time.Since(started).Milliseconds()
		r.ErrorClass = classifyOpenAIStatus(resp.StatusCode)
		return r, nil
	}

	durationMs, gotToken := scanOpenAIStream(bufio.NewScanner(resp.Body), started)
	r.DurationMs = durationMs
	if !gotToken {
		r.ErrorClass = probes.ErrorClassMalformedBody
		r.ErrorDetail = "stream ended without content token"
		return r, nil
	}

	r.Success = true
	return r, nil
}

// scanOpenAIStream reads SSE lines until a non-empty content token is found.
// Returns the elapsed ms and whether a token was received.
func scanOpenAIStream(scanner *bufio.Scanner, started time.Time) (int64, bool) {
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}
		var chunk openaiStreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			return time.Since(started).Milliseconds(), true
		}
	}
	return time.Since(started).Milliseconds(), false
}
