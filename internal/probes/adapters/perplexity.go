package adapters

import (
	"context"
	"net/http"
	"time"

	"github.com/llmstatus/llmstatus/internal/httpclient"
	"github.com/llmstatus/llmstatus/internal/probes"
)

const (
	perplexityDefaultBaseURL = "https://api.perplexity.ai"
	perplexityProviderID     = "perplexity"
	perplexityLightModel     = "sonar"
	perplexityLightProbeType = "light_inference"
	perplexityErrorDetailMax = 200
)

// PerplexityOption configures a Perplexity provider at construction time.
type PerplexityOption func(*perplexityProvider)

// WithPerplexityBaseURL overrides the base URL. Intended for tests.
func WithPerplexityBaseURL(u string) PerplexityOption {
	return func(p *perplexityProvider) { p.baseURL = u }
}

// WithPerplexityHTTPClient overrides the HTTP client. Intended for tests.
func WithPerplexityHTTPClient(c *http.Client) PerplexityOption {
	return func(p *perplexityProvider) { p.client = c }
}

// NewPerplexityProvider returns a probes.Provider backed by api.perplexity.ai.
func NewPerplexityProvider(apiKey, region string, opts ...PerplexityOption) probes.Provider {
	p := &perplexityProvider{
		baseURL: perplexityDefaultBaseURL,
		apiKey:  apiKey,
		region:  region,
		client:  httpclient.New(httpclient.Options{Timeout: 30 * time.Second}),
	}
	for _, o := range opts {
		o(p)
	}
	return p
}

type perplexityProvider struct {
	baseURL, apiKey, region string
	client                  *http.Client
}

func (p *perplexityProvider) ID() string       { return perplexityProviderID }
func (p *perplexityProvider) Models() []string { return []string{perplexityLightModel} }

func (p *perplexityProvider) ProbeLightInference(ctx context.Context, model string) (probes.ProbeResult, error) {
	return probeOpenAICompat(ctx, p.baseURL, p.apiKey, p.region, perplexityProviderID, perplexityLightProbeType, perplexityErrorDetailMax, model, p.client)
}

func (p *perplexityProvider) ProbeQuality(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: perplexityProviderID, ProbeType: "quality"}
}
func (p *perplexityProvider) ProbeEmbedding(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: perplexityProviderID, ProbeType: "embedding"}
}
func (p *perplexityProvider) ProbeStreaming(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: perplexityProviderID, ProbeType: "streaming"}
}
