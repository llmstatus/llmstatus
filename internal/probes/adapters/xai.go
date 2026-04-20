package adapters

import (
	"context"
	"net/http"
	"time"

	"github.com/llmstatus/llmstatus/internal/httpclient"
	"github.com/llmstatus/llmstatus/internal/probes"
)

const (
	xaiDefaultBaseURL = "https://api.x.ai/v1"
	xaiProviderID     = "xai"
	xaiLightModel     = "grok-3-mini"
	xaiLightProbeType = "light_inference"
	xaiErrorDetailMax = 200
)

// XAIOption configures an xAI provider at construction time.
type XAIOption func(*xaiProvider)

// WithXAIBaseURL overrides the base URL. Intended for tests.
func WithXAIBaseURL(u string) XAIOption { return func(p *xaiProvider) { p.baseURL = u } }

// WithXAIHTTPClient overrides the HTTP client. Intended for tests.
func WithXAIHTTPClient(c *http.Client) XAIOption { return func(p *xaiProvider) { p.client = c } }

// NewXAIProvider returns a probes.Provider backed by api.x.ai (Grok).
func NewXAIProvider(apiKey, region string, opts ...XAIOption) probes.Provider {
	p := &xaiProvider{
		baseURL: xaiDefaultBaseURL,
		apiKey:  apiKey,
		region:  region,
		client:  httpclient.New(httpclient.Options{Timeout: 30 * time.Second}),
	}
	for _, o := range opts {
		o(p)
	}
	return p
}

type xaiProvider struct {
	baseURL, apiKey, region string
	client                  *http.Client
}

func (p *xaiProvider) ID() string       { return xaiProviderID }
func (p *xaiProvider) Models() []string { return []string{xaiLightModel} }

func (p *xaiProvider) ProbeLightInference(ctx context.Context, model string) (probes.ProbeResult, error) {
	return probeOpenAICompat(ctx, p.baseURL, p.apiKey, p.region, xaiProviderID, xaiLightProbeType, xaiErrorDetailMax, model, p.client)
}

func (p *xaiProvider) ProbeQuality(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: xaiProviderID, ProbeType: "quality"}
}
func (p *xaiProvider) ProbeEmbedding(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: xaiProviderID, ProbeType: "embedding"}
}
func (p *xaiProvider) ProbeStreaming(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: xaiProviderID, ProbeType: "streaming"}
}
