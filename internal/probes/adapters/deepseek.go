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
	deepseekDefaultBaseURL = "https://api.deepseek.com/v1"
	deepseekProviderID     = "deepseek"
	deepseekLightModel     = "deepseek-chat"
	deepseekLightProbeType = "light_inference"
	deepseekErrorDetailMax = 200
)

// DeepSeekOption configures a DeepSeek provider at construction time.
type DeepSeekOption func(*deepseekProvider)

// WithDeepSeekBaseURL overrides the base URL. Intended for tests.
func WithDeepSeekBaseURL(u string) DeepSeekOption {
	return func(p *deepseekProvider) { p.baseURL = u }
}

// WithDeepSeekHTTPClient overrides the HTTP client. Intended for tests.
func WithDeepSeekHTTPClient(c *http.Client) DeepSeekOption {
	return func(p *deepseekProvider) { p.client = c }
}

// NewDeepSeekProvider returns a probes.Provider backed by api.deepseek.com.
// DeepSeek exposes an OpenAI-compatible Chat Completions API.
func NewDeepSeekProvider(apiKey, region string, opts ...DeepSeekOption) probes.Provider {
	p := &deepseekProvider{
		baseURL: deepseekDefaultBaseURL,
		apiKey:  apiKey,
		region:  region,
		client:  httpclient.New(httpclient.Options{Timeout: 30 * time.Second}),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

type deepseekProvider struct {
	baseURL string
	apiKey  string
	region  string
	client  *http.Client
}

func (p *deepseekProvider) ID() string       { return deepseekProviderID }
func (p *deepseekProvider) Models() []string { return []string{deepseekLightModel} }

func (p *deepseekProvider) ProbeLightInference(ctx context.Context, model string) (probes.ProbeResult, error) {
	started := time.Now()
	r := probes.ProbeResult{
		ProviderID: deepseekProviderID,
		Model:      model,
		ProbeType:  deepseekLightProbeType,
		StartedAt:  started.UTC(),
		RegionID:   p.region,
	}

	req, err := p.buildRequest(ctx, model)
	if err != nil {
		r.DurationMs = time.Since(started).Milliseconds()
		r.ErrorClass = probes.ErrorClassUnknown
		r.ErrorDetail = truncate(err.Error(), deepseekErrorDetailMax)
		return r, err
	}

	resp, err := p.client.Do(req)
	r.DurationMs = time.Since(started).Milliseconds()
	if err != nil {
		r.ErrorClass = classifyNetError(err)
		r.ErrorDetail = truncate(err.Error(), deepseekErrorDetailMax)
		return r, nil
	}
	defer func() { _ = resp.Body.Close() }()

	r.HTTPStatus = resp.StatusCode
	body, _ := io.ReadAll(resp.Body)
	// DeepSeek uses the same response envelope as OpenAI.
	classifyOpenAIResponse(&r, resp.StatusCode, body)
	return r, nil
}

func (p *deepseekProvider) ProbeQuality(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: deepseekProviderID, ProbeType: "quality"}
}

func (p *deepseekProvider) ProbeEmbedding(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: deepseekProviderID, ProbeType: "embedding"}
}

func (p *deepseekProvider) ProbeStreaming(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: deepseekProviderID, ProbeType: "streaming"}
}

func (p *deepseekProvider) buildRequest(ctx context.Context, model string) (*http.Request, error) {
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
	return req, nil
}
