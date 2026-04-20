package adapters

import (
	"context"
	"net/http"
	"time"

	"github.com/llmstatus/llmstatus/internal/httpclient"
	"github.com/llmstatus/llmstatus/internal/probes"
)

const (
	togetherDefaultBaseURL = "https://api.together.xyz/v1"
	togetherProviderID     = "together_ai"
	togetherLightModel     = "meta-llama/Llama-3-8b-chat-hf"
	togetherLightProbeType = "light_inference"
	togetherErrorDetailMax = 200
)

// TogetherOption configures a Together AI provider at construction time.
type TogetherOption func(*togetherProvider)

// WithTogetherBaseURL overrides the base URL. Intended for tests.
func WithTogetherBaseURL(u string) TogetherOption { return func(p *togetherProvider) { p.baseURL = u } }

// WithTogetherHTTPClient overrides the HTTP client. Intended for tests.
func WithTogetherHTTPClient(c *http.Client) TogetherOption {
	return func(p *togetherProvider) { p.client = c }
}

// NewTogetherProvider returns a probes.Provider backed by api.together.xyz.
func NewTogetherProvider(apiKey, region string, opts ...TogetherOption) probes.Provider {
	p := &togetherProvider{
		baseURL: togetherDefaultBaseURL,
		apiKey:  apiKey,
		region:  region,
		client:  httpclient.New(httpclient.Options{Timeout: 30 * time.Second}),
	}
	for _, o := range opts {
		o(p)
	}
	return p
}

type togetherProvider struct {
	baseURL, apiKey, region string
	client                  *http.Client
}

func (p *togetherProvider) ID() string       { return togetherProviderID }
func (p *togetherProvider) Models() []string { return []string{togetherLightModel} }

func (p *togetherProvider) ProbeLightInference(ctx context.Context, model string) (probes.ProbeResult, error) {
	return probeOpenAICompat(ctx, p.baseURL, p.apiKey, p.region, togetherProviderID, togetherLightProbeType, togetherErrorDetailMax, model, p.client)
}

func (p *togetherProvider) ProbeQuality(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: togetherProviderID, ProbeType: "quality"}
}
func (p *togetherProvider) ProbeEmbedding(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: togetherProviderID, ProbeType: "embedding"}
}
func (p *togetherProvider) ProbeStreaming(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: togetherProviderID, ProbeType: "streaming"}
}
