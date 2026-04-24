package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/llmstatus/llmstatus/internal/probes"
)

const (
	anthropicQualityProbeType = "quality"
	anthropicQualityPrompt    = "Reply with only the single word PONG and nothing else."
	anthropicQualityMaxTokens = 5
	anthropicQualityExpected  = "PONG"
)

// ProbeQuality sends a fixed prompt and verifies the model echoes back the
// expected token. A 200 response whose content does not contain the expected
// word is classified as ErrorClassQualityMismatch.
func (p *anthropicProvider) ProbeQuality(ctx context.Context, model string) (probes.ProbeResult, error) {
	started := time.Now()
	r := probes.ProbeResult{
		ProviderID: anthropicProviderID,
		Model:      model,
		ProbeType:  anthropicQualityProbeType,
		StartedAt:  started.UTC(),
		RegionID:   p.region,
	}

	body, err := json.Marshal(anthropicMessagesRequest{
		Model:     model,
		MaxTokens: anthropicQualityMaxTokens,
		Messages:  []anthropicMessage{{Role: "user", Content: anthropicQualityPrompt}},
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

	resp, err := p.client.Do(req)
	r.DurationMs = time.Since(started).Milliseconds()
	if err != nil {
		r.ErrorClass = classifyNetError(err)
		r.ErrorDetail = truncate(err.Error(), anthropicErrorDetailMax)
		return r, nil
	}
	defer func() { _ = resp.Body.Close() }()

	r.HTTPStatus = resp.StatusCode
	rawBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		r.ErrorClass = classifyAnthropicStatus(resp.StatusCode)
		r.ErrorDetail = parseAnthropicError(rawBody)
		return r, nil
	}

	var cr anthropicMessagesResponse
	if err := json.Unmarshal(rawBody, &cr); err != nil || len(cr.Content) == 0 {
		r.ErrorClass = probes.ErrorClassMalformedBody
		r.ErrorDetail = truncate(string(rawBody), anthropicErrorDetailMax)
		return r, nil
	}

	r.TokensIn = cr.Usage.InputTokens
	r.TokensOut = cr.Usage.OutputTokens

	content := strings.TrimSpace(cr.Content[0].Text)
	if strings.Contains(strings.ToUpper(content), anthropicQualityExpected) {
		r.Success = true
	} else {
		r.ErrorClass = probes.ErrorClassQualityMismatch
		r.ErrorDetail = truncate("got: "+content, anthropicErrorDetailMax)
	}
	return r, nil
}
