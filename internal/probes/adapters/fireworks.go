package adapters

import (
	"context"
	"net/http"
	"time"

	"github.com/llmstatus/llmstatus/internal/httpclient"
	"github.com/llmstatus/llmstatus/internal/probes"
)

const (
	fireworksDefaultBaseURL = "https://api.fireworks.ai/inference/v1"
	fireworksProviderID     = "fireworks"
	fireworksLightModel     = "accounts/fireworks/models/llama-v3p1-8b-instruct"
	fireworksLightProbeType = "light_inference"
	fireworksErrorDetailMax = 200
)

// NewFireworksProviderOption configures a NewFireworksProvider provider at construction time.
type NewFireworksProviderOption func(*fireworksProvider)

// WithNewFireworksProviderBaseURL overrides the base URL. Intended for tests.
func WithNewFireworksProviderBaseURL(u string) NewFireworksProviderOption { return func(p *fireworksProvider) { p.baseURL = u } }

// WithNewFireworksProviderHTTPClient overrides the HTTP client. Intended for tests.
func WithNewFireworksProviderHTTPClient(c *http.Client) NewFireworksProviderOption {
	return func(p *fireworksProvider) { p.client = c }
}

// NewFireworksProvider returns a probes.Provider backed by https://api.fireworks.ai/inference/v1.
func NewFireworksProvider(apiKey, region string, opts ...NewFireworksProviderOption) probes.Provider {
	p := &fireworksProvider{
		baseURL: fireworksDefaultBaseURL,
		apiKey:  apiKey,
		region:  region,
		client:  httpclient.New(httpclient.Options{Timeout: 30 * time.Second}),
	}
	for _, o := range opts {
		o(p)
	}
	return p
}

type fireworksProvider struct {
	baseURL, apiKey, region string
	client                  *http.Client
}

func (p *fireworksProvider) ID() string       { return fireworksProviderID }
func (p *fireworksProvider) Models() []string { return []string{fireworksLightModel} }

func (p *fireworksProvider) ProbeLightInference(ctx context.Context, model string) (probes.ProbeResult, error) {
	return probeOpenAICompat(ctx, p.baseURL, p.apiKey, p.region, fireworksProviderID, fireworksLightProbeType, fireworksErrorDetailMax, model, p.client)
}

func (p *fireworksProvider) ProbeQuality(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: fireworksProviderID, ProbeType: "quality"}
}
func (p *fireworksProvider) ProbeEmbedding(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: fireworksProviderID, ProbeType: "embedding"}
}
func (p *fireworksProvider) ProbeStreaming(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: fireworksProviderID, ProbeType: "streaming"}
}
