package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/llmstatus/llmstatus/internal/probes"
)

const (
	openaiEmbeddingProbeType = "embedding"
	// openaiEmbeddingModel is always used regardless of the model argument;
	// embeddings are a provider-level capability, not per-chat-model.
	openaiEmbeddingModel = "text-embedding-3-small"
	openaiEmbeddingInput = "hello"
)

type openaiEmbeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type openaiEmbeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// ProbeEmbedding issues a minimal embeddings request and verifies the response
// contains a non-empty vector. The model argument is ignored; text-embedding-3-small
// is always used because embeddings are a provider-level capability.
func (p *openaiProvider) ProbeEmbedding(ctx context.Context, _ string) (probes.ProbeResult, error) {
	started := time.Now()
	r := probes.ProbeResult{
		ProviderID: openaiProviderID,
		Model:      openaiEmbeddingModel,
		ProbeType:  openaiEmbeddingProbeType,
		StartedAt:  started.UTC(),
		RegionID:   p.region,
	}

	body, err := json.Marshal(openaiEmbeddingRequest{
		Model: openaiEmbeddingModel,
		Input: openaiEmbeddingInput,
	})
	if err != nil {
		r.DurationMs = time.Since(started).Milliseconds()
		r.ErrorClass = probes.ErrorClassUnknown
		r.ErrorDetail = truncate(err.Error(), openaiErrorDetailMax)
		return r, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/embeddings", bytes.NewReader(body))
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

	var er openaiEmbeddingResponse
	if err := json.Unmarshal(rawBody, &er); err != nil || len(er.Data) == 0 || len(er.Data[0].Embedding) == 0 {
		r.ErrorClass = probes.ErrorClassMalformedBody
		r.ErrorDetail = truncate(string(rawBody), openaiErrorDetailMax)
		return r, nil
	}

	r.Success = true
	r.TokensIn = er.Usage.PromptTokens
	return r, nil
}
