package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/llmstatus/llmstatus/internal/httpclient"
	"github.com/llmstatus/llmstatus/internal/probes"
)

const (
	anthropicDefaultBaseURL  = "https://api.anthropic.com/v1"
	anthropicProviderID      = "anthropic"
	anthropicVersion         = "2023-06-01"
	anthropicLightModel      = "claude-haiku-4-5-20251001"
	anthropicLightProbeType  = "light_inference"
	anthropicErrorDetailMax  = 200
	// 529 is Anthropic's non-standard "overloaded" HTTP status (see docs/known-quirks.md).
	anthropicStatusOverloaded = 529
)

func init() {
	// Register a zero-key placeholder so the adapter appears in the registry.
	// cmd/prober replaces this with a keyed instance loaded from the environment.
	Register(NewAnthropicProvider("", ""))
}

// AnthropicOption configures an Anthropic provider at construction time.
type AnthropicOption func(*anthropicProvider)

// WithAnthropicBaseURL overrides the API base URL. Used in tests.
func WithAnthropicBaseURL(u string) AnthropicOption {
	return func(p *anthropicProvider) { p.baseURL = u }
}

// WithAnthropicHTTPClient overrides the HTTP client. Used in tests.
func WithAnthropicHTTPClient(c *http.Client) AnthropicOption {
	return func(p *anthropicProvider) { p.client = c }
}

// NewAnthropicProvider returns a probes.Provider backed by api.anthropic.com.
func NewAnthropicProvider(apiKey, region string, opts ...AnthropicOption) probes.Provider {
	p := &anthropicProvider{
		baseURL: anthropicDefaultBaseURL,
		apiKey:  apiKey,
		region:  region,
		client:  httpclient.New(httpclient.Options{Timeout: 30 * time.Second}),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

type anthropicProvider struct {
	baseURL string
	apiKey  string
	region  string
	client  *http.Client
}

func (p *anthropicProvider) ID() string       { return anthropicProviderID }
func (p *anthropicProvider) Models() []string { return []string{anthropicLightModel} }

func (p *anthropicProvider) ProbeLightInference(ctx context.Context, model string) (probes.ProbeResult, error) {
	started := time.Now()
	r := probes.ProbeResult{
		ProviderID: anthropicProviderID,
		Model:      model,
		ProbeType:  anthropicLightProbeType,
		StartedAt:  started.UTC(),
		RegionID:   p.region,
	}

	req, err := p.buildLightRequest(ctx, model)
	if err != nil {
		r.DurationMs = time.Since(started).Milliseconds()
		r.ErrorClass = probes.ErrorClassUnknown
		r.ErrorDetail = truncate(err.Error(), anthropicErrorDetailMax)
		return r, err
	}

	resp, err := p.client.Do(req)
	r.DurationMs = time.Since(started).Milliseconds()
	if err != nil {
		r.ErrorClass = classifyNetError(err)
		r.ErrorDetail = truncate(err.Error(), anthropicErrorDetailMax)
		return r, nil
	}
	defer func() { _ = resp.Body.Close() }()

	r.HTTPStatus = resp.StatusCode
	body, _ := io.ReadAll(resp.Body)
	classifyAnthropicResponse(&r, resp.StatusCode, body)
	return r, nil
}

func (p *anthropicProvider) ProbeQuality(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: anthropicProviderID, ProbeType: "quality"}
}

func (p *anthropicProvider) ProbeEmbedding(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: anthropicProviderID, ProbeType: "embedding"}
}

func (p *anthropicProvider) ProbeStreaming(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: anthropicProviderID, ProbeType: "streaming"}
}

// ---- request / response types -----------------------------------------------

type anthropicMessagesRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicMessagesResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

type anthropicErrorEnvelope struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// ---- helpers ----------------------------------------------------------------

func (p *anthropicProvider) buildLightRequest(ctx context.Context, model string) (*http.Request, error) {
	body, err := json.Marshal(anthropicMessagesRequest{
		Model:     model,
		MaxTokens: 1,
		Messages:  []anthropicMessage{{Role: "user", Content: "ping"}},
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", anthropicVersion)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func classifyAnthropicResponse(r *probes.ProbeResult, status int, body []byte) {
	if status >= 200 && status < 300 {
		var cr anthropicMessagesResponse
		if err := json.Unmarshal(body, &cr); err != nil {
			r.ErrorClass = probes.ErrorClassMalformedBody
			r.ErrorDetail = truncate(string(body), anthropicErrorDetailMax)
			return
		}
		r.Success = true
		r.TokensIn = cr.Usage.InputTokens
		r.TokensOut = cr.Usage.OutputTokens
		return
	}
	r.ErrorClass = classifyAnthropicStatus(status)
	r.ErrorDetail = parseAnthropicError(body)
}

func classifyAnthropicStatus(status int) probes.ErrorClass {
	switch {
	case status == http.StatusUnauthorized, status == http.StatusForbidden:
		return probes.ErrorClassAuth
	case status == http.StatusTooManyRequests:
		return probes.ErrorClassRateLimit
	case status == anthropicStatusOverloaded, status >= 500:
		return probes.ErrorClassServer5xx
	case status >= 400:
		return probes.ErrorClassClient4xx
	default:
		return probes.ErrorClassUnknown
	}
}

func parseAnthropicError(body []byte) string {
	var e anthropicErrorEnvelope
	if err := json.Unmarshal(body, &e); err == nil && e.Error.Message != "" {
		return truncate(e.Error.Message, anthropicErrorDetailMax)
	}
	return truncate(string(body), anthropicErrorDetailMax)
}
