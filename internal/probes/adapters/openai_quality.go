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
	openaiQualityProbeType = "quality"
	openaiQualityPrompt    = "Reply with only the single word PONG and nothing else."
	openaiQualityMaxTokens = 5
	openaiQualityExpected  = "PONG"
)

// ProbeQuality sends a fixed prompt and verifies the model echoes back the
// expected token. A 200 response whose content does not contain the expected
// word is classified as ErrorClassQualityMismatch.
func (p *openaiProvider) ProbeQuality(ctx context.Context, model string) (probes.ProbeResult, error) {
	if isOpenAIEmbeddingModel(model) {
		return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: openaiProviderID, ProbeType: "quality"}
	}
	started := time.Now()
	r := probes.ProbeResult{
		ProviderID: openaiProviderID,
		Model:      model,
		ProbeType:  openaiQualityProbeType,
		StartedAt:  started.UTC(),
		RegionID:   p.region,
	}

	body, err := json.Marshal(openaiChatRequest{
		Model:     model,
		Messages:  []openaiChatMessage{{Role: "user", Content: openaiQualityPrompt}},
		MaxTokens: openaiQualityMaxTokens,
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

	resp, err := p.client.Do(req)
	r.DurationMs = time.Since(started).Milliseconds()
	if err != nil {
		r.ErrorClass = classifyNetError(err)
		r.ErrorDetail = truncate(err.Error(), openaiErrorDetailMax)
		return r, nil
	}
	defer func() { _ = resp.Body.Close() }()

	r.HTTPStatus = resp.StatusCode
	rawBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		r.ErrorClass = classifyOpenAIStatus(resp.StatusCode)
		r.ErrorDetail = parseOpenAIError(rawBody)
		return r, nil
	}

	var cr openaiChatResponse
	if err := json.Unmarshal(rawBody, &cr); err != nil || len(cr.Choices) == 0 {
		r.ErrorClass = probes.ErrorClassMalformedBody
		r.ErrorDetail = truncate(string(rawBody), openaiErrorDetailMax)
		return r, nil
	}

	r.TokensIn = cr.Usage.PromptTokens
	r.TokensOut = cr.Usage.CompletionTokens

	content := strings.TrimSpace(cr.Choices[0].Message.Content)
	if strings.Contains(strings.ToUpper(content), openaiQualityExpected) {
		r.Success = true
	} else {
		r.ErrorClass = probes.ErrorClassQualityMismatch
		r.ErrorDetail = truncate("got: "+content, openaiErrorDetailMax)
	}
	return r, nil
}
