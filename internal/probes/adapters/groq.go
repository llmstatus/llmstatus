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
	groqDefaultBaseURL = "https://api.groq.com/openai/v1"
	groqProviderID     = "groq"
	groqLightModel     = "llama-3.3-70b-versatile"
	groqLightProbeType = "light_inference"
	groqErrorDetailMax = 200
)

// GroqOption configures a Groq provider at construction time.
type GroqOption func(*groqProvider)

// WithGroqBaseURL overrides the base URL. Intended for tests.
func WithGroqBaseURL(u string) GroqOption { return func(p *groqProvider) { p.baseURL = u } }

// WithGroqHTTPClient overrides the HTTP client. Intended for tests.
func WithGroqHTTPClient(c *http.Client) GroqOption { return func(p *groqProvider) { p.client = c } }

// NewGroqProvider returns a probes.Provider backed by api.groq.com.
func NewGroqProvider(apiKey, region string, opts ...GroqOption) probes.Provider {
	p := &groqProvider{
		baseURL: groqDefaultBaseURL,
		apiKey:  apiKey,
		region:  region,
		client:  httpclient.New(httpclient.Options{Timeout: 30 * time.Second}),
	}
	for _, o := range opts {
		o(p)
	}
	return p
}

type groqProvider struct {
	baseURL, apiKey, region string
	client                  *http.Client
}

func (p *groqProvider) ID() string       { return groqProviderID }
func (p *groqProvider) Models() []string { return []string{groqLightModel} }

func (p *groqProvider) ProbeLightInference(ctx context.Context, model string) (probes.ProbeResult, error) {
	return probeOpenAICompat(ctx, p.baseURL, p.apiKey, p.region, groqProviderID, groqLightProbeType, groqErrorDetailMax, model, p.client)
}

func (p *groqProvider) ProbeQuality(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: groqProviderID, ProbeType: "quality"}
}
func (p *groqProvider) ProbeEmbedding(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: groqProviderID, ProbeType: "embedding"}
}
func (p *groqProvider) ProbeStreaming(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: groqProviderID, ProbeType: "streaming"}
}

// probeOpenAICompat is the shared probe implementation for all providers that
// expose the OpenAI Chat Completions API at /chat/completions.
func probeOpenAICompat(ctx context.Context, baseURL, apiKey, region, providerID, probeType string, maxDetail int, model string, client *http.Client) (probes.ProbeResult, error) {
	started := time.Now()
	r := probes.ProbeResult{
		ProviderID: providerID,
		Model:      model,
		ProbeType:  probeType,
		StartedAt:  started.UTC(),
		RegionID:   region,
	}

	payload, err := json.Marshal(openaiChatRequest{
		Model:     model,
		Messages:  []openaiChatMessage{{Role: "user", Content: "ping"}},
		MaxTokens: 1,
	})
	if err != nil {
		r.DurationMs = time.Since(started).Milliseconds()
		r.ErrorClass = probes.ErrorClassUnknown
		r.ErrorDetail = truncate(err.Error(), maxDetail)
		return r, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		r.DurationMs = time.Since(started).Milliseconds()
		r.ErrorClass = probes.ErrorClassUnknown
		r.ErrorDetail = truncate(err.Error(), maxDetail)
		return r, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	r.DurationMs = time.Since(started).Milliseconds()
	if err != nil {
		r.ErrorClass = classifyNetError(err)
		r.ErrorDetail = truncate(err.Error(), maxDetail)
		return r, nil
	}
	defer func() { _ = resp.Body.Close() }()

	r.HTTPStatus = resp.StatusCode
	body, _ := io.ReadAll(resp.Body)
	classifyOpenAIResponse(&r, resp.StatusCode, body)
	return r, nil
}
