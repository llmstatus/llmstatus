package adapters

import (
	"context"
	"net/http"
	"time"

	"github.com/llmstatus/llmstatus/internal/httpclient"
	"github.com/llmstatus/llmstatus/internal/probes"
)

const (
	minimaxDefaultBaseURL = "https://api.minimax.chat/v1"
	minimaxProviderID     = "minimax"
	minimaxLightModel     = "abab6.5s-chat"
	minimaxLightProbeType = "light_inference"
	minimaxErrorDetailMax = 200
)

// NewMinimaxProviderOption configures a NewMinimaxProvider provider at construction time.
type NewMinimaxProviderOption func(*minimaxProvider)

// WithNewMinimaxProviderBaseURL overrides the base URL. Intended for tests.
func WithNewMinimaxProviderBaseURL(u string) NewMinimaxProviderOption {
	return func(p *minimaxProvider) { p.baseURL = u }
}

// WithNewMinimaxProviderHTTPClient overrides the HTTP client. Intended for tests.
func WithNewMinimaxProviderHTTPClient(c *http.Client) NewMinimaxProviderOption {
	return func(p *minimaxProvider) { p.client = c }
}

// NewMinimaxProvider returns a probes.Provider backed by https://api.minimax.chat/v1.
func NewMinimaxProvider(apiKey, region string, opts ...NewMinimaxProviderOption) probes.Provider {
	p := &minimaxProvider{
		baseURL: minimaxDefaultBaseURL,
		apiKey:  apiKey,
		region:  region,
		client:  httpclient.New(httpclient.Options{Timeout: 30 * time.Second}),
	}
	for _, o := range opts {
		o(p)
	}
	return p
}

type minimaxProvider struct {
	baseURL, apiKey, region string
	client                  *http.Client
}

func (p *minimaxProvider) ID() string       { return minimaxProviderID }
func (p *minimaxProvider) Models() []string { return []string{minimaxLightModel} }

func (p *minimaxProvider) ProbeLightInference(ctx context.Context, model string) (probes.ProbeResult, error) {
	return probeOpenAICompat(ctx, p.baseURL, p.apiKey, p.region, minimaxProviderID, minimaxLightProbeType, minimaxErrorDetailMax, model, p.client)
}

func (p *minimaxProvider) ProbeQuality(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: minimaxProviderID, ProbeType: "quality"}
}
func (p *minimaxProvider) ProbeEmbedding(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: minimaxProviderID, ProbeType: "embedding"}
}
func (p *minimaxProvider) ProbeStreaming(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: minimaxProviderID, ProbeType: "streaming"}
}
