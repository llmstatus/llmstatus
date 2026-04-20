package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/llmstatus/llmstatus/internal/httpclient"
	"github.com/llmstatus/llmstatus/internal/probes"
)

const (
	openaiDefaultBaseURL = "https://api.openai.com/v1"
	openaiProviderID     = "openai"
	openaiLightModel     = "gpt-4o-mini"
	openaiLightProbeType = "light_inference"
	openaiErrorDetailMax = 200

	// probeUserAgent is sent on every outbound probe request so providers can
	// identify monitoring traffic and we are transparent about our purpose.
	probeUserAgent = "llmstatus.io/1.0 (+https://llmstatus.io)"
)

// OpenAIOption configures an OpenAI provider at construction time.
type OpenAIOption func(*openaiProvider)

// WithOpenAIBaseURL overrides the base URL. Intended for tests; production
// callers should not need this.
func WithOpenAIBaseURL(u string) OpenAIOption {
	return func(p *openaiProvider) { p.baseURL = u }
}

// WithOpenAIHTTPClient overrides the HTTP client. Intended for tests and
// for injecting a differently-configured client (for example, with a
// tighter timeout) in special probe scenarios.
func WithOpenAIHTTPClient(c *http.Client) OpenAIOption {
	return func(p *openaiProvider) { p.client = c }
}

// NewOpenAIProvider returns a probes.Provider backed by api.openai.com.
//
// region is copied into every ProbeResult.RegionID produced by this
// provider; probe runners supply the node identifier when constructing
// one adapter per node.
func NewOpenAIProvider(apiKey, region string, opts ...OpenAIOption) probes.Provider {
	p := &openaiProvider{
		baseURL: openaiDefaultBaseURL,
		apiKey:  apiKey,
		region:  region,
		client:  httpclient.New(httpclient.Options{Timeout: 30 * time.Second}),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

type openaiProvider struct {
	baseURL string
	apiKey  string
	region  string
	client  *http.Client
}

func (p *openaiProvider) ID() string       { return openaiProviderID }
func (p *openaiProvider) Models() []string { return []string{openaiLightModel} }

// ProbeLightInference sends a minimal chat completion and classifies the
// response into a probes.ProbeResult. It never returns a non-nil error
// for expected provider failures (timeout, rate-limit, 5xx, auth,
// malformed body): those are carried in the result's ErrorClass.
// It only returns a non-nil error for programmer / environment bugs
// (context cancelled, marshalling failure).
func (p *openaiProvider) ProbeLightInference(ctx context.Context, model string) (probes.ProbeResult, error) {
	started := time.Now()
	r := probes.ProbeResult{
		ProviderID: openaiProviderID,
		Model:      model,
		ProbeType:  openaiLightProbeType,
		StartedAt:  started.UTC(),
		RegionID:   p.region,
	}

	req, err := p.buildLightRequest(ctx, model)
	if err != nil {
		r.DurationMs = time.Since(started).Milliseconds()
		r.ErrorClass = probes.ErrorClassUnknown
		r.ErrorDetail = truncate(err.Error(), openaiErrorDetailMax)
		return r, err
	}

	resp, err := p.client.Do(req)
	r.DurationMs = time.Since(started).Milliseconds()
	if err != nil {
		r.ErrorClass = classifyNetError(err)
		r.ErrorDetail = truncate(err.Error(), openaiErrorDetailMax)
		return r, nil
	}
	defer func() { _ = resp.Body.Close() }()

	r.HTTPStatus = resp.StatusCode
	body, _ := io.ReadAll(resp.Body)
	classifyOpenAIResponse(&r, resp.StatusCode, body)
	return r, nil
}

// ProbeQuality, ProbeEmbedding, ProbeStreaming are in openai_quality.go,
// openai_embedding.go, openai_streaming.go respectively.

type openaiChatRequest struct {
	Model     string              `json:"model"`
	Messages  []openaiChatMessage `json:"messages"`
	MaxTokens int                 `json:"max_tokens"`
}

type openaiChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiChatResponse struct {
	Choices []struct {
		Message openaiChatMessage `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

type openaiErrorEnvelope struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code,omitempty"`
	} `json:"error"`
}

func (p *openaiProvider) buildLightRequest(ctx context.Context, model string) (*http.Request, error) {
	body, err := json.Marshal(openaiChatRequest{
		Model:     model,
		Messages:  []openaiChatMessage{{Role: "user", Content: "ping"}},
		MaxTokens: 1,
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", probeUserAgent)
	return req, nil
}

func classifyOpenAIResponse(r *probes.ProbeResult, status int, body []byte) {
	if status >= 200 && status < 300 {
		var cr openaiChatResponse
		if err := json.Unmarshal(body, &cr); err != nil {
			r.ErrorClass = probes.ErrorClassMalformedBody
			r.ErrorDetail = truncate(string(body), openaiErrorDetailMax)
			return
		}
		r.Success = true
		r.TokensIn = cr.Usage.PromptTokens
		r.TokensOut = cr.Usage.CompletionTokens
		return
	}
	r.ErrorClass = classifyOpenAIStatus(status)
	r.ErrorDetail = parseOpenAIError(body)
}

func classifyOpenAIStatus(status int) probes.ErrorClass {
	switch {
	case status == http.StatusUnauthorized, status == http.StatusForbidden:
		return probes.ErrorClassAuth
	case status == http.StatusTooManyRequests:
		return probes.ErrorClassRateLimit
	case status >= 500:
		return probes.ErrorClassServer5xx
	case status >= 400:
		return probes.ErrorClassClient4xx
	default:
		return probes.ErrorClassUnknown
	}
}

func classifyNetError(err error) probes.ErrorClass {
	// Prefer structured checks over string matching when possible.
	if errors.Is(err, context.DeadlineExceeded) {
		return probes.ErrorClassTimeout
	}
	msg := err.Error()
	if strings.Contains(msg, "deadline exceeded") ||
		strings.Contains(msg, "Client.Timeout") ||
		strings.Contains(msg, "i/o timeout") {
		return probes.ErrorClassTimeout
	}
	return probes.ErrorClassUnknown
}

func parseOpenAIError(body []byte) string {
	var e openaiErrorEnvelope
	if err := json.Unmarshal(body, &e); err == nil && e.Error.Message != "" {
		return truncate(e.Error.Message, openaiErrorDetailMax)
	}
	return truncate(string(body), openaiErrorDetailMax)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
