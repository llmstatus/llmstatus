package adapters

import (
	"context"
	"net/http"
	"time"

	"github.com/llmstatus/llmstatus/internal/httpclient"
	"github.com/llmstatus/llmstatus/internal/probes"
)

const (
	moonshotDefaultBaseURL = "https://api.moonshot.cn/v1"
	moonshotProviderID     = "moonshot"
	moonshotLightModel     = "moonshot-v1-8k"
	moonshotLightProbeType = "light_inference"
	moonshotErrorDetailMax = 200
)

// NewMoonshotProviderOption configures a NewMoonshotProvider provider at construction time.
type NewMoonshotProviderOption func(*moonshotProvider)

// WithNewMoonshotProviderBaseURL overrides the base URL. Intended for tests.
func WithNewMoonshotProviderBaseURL(u string) NewMoonshotProviderOption { return func(p *moonshotProvider) { p.baseURL = u } }

// WithNewMoonshotProviderHTTPClient overrides the HTTP client. Intended for tests.
func WithNewMoonshotProviderHTTPClient(c *http.Client) NewMoonshotProviderOption {
	return func(p *moonshotProvider) { p.client = c }
}

// NewMoonshotProvider returns a probes.Provider backed by https://api.moonshot.cn/v1.
func NewMoonshotProvider(apiKey, region string, opts ...NewMoonshotProviderOption) probes.Provider {
	p := &moonshotProvider{
		baseURL: moonshotDefaultBaseURL,
		apiKey:  apiKey,
		region:  region,
		client:  httpclient.New(httpclient.Options{Timeout: 30 * time.Second}),
	}
	for _, o := range opts {
		o(p)
	}
	return p
}

type moonshotProvider struct {
	baseURL, apiKey, region string
	client                  *http.Client
}

func (p *moonshotProvider) ID() string       { return moonshotProviderID }
func (p *moonshotProvider) Models() []string { return []string{moonshotLightModel} }

func (p *moonshotProvider) ProbeLightInference(ctx context.Context, model string) (probes.ProbeResult, error) {
	return probeOpenAICompat(ctx, p.baseURL, p.apiKey, p.region, moonshotProviderID, moonshotLightProbeType, moonshotErrorDetailMax, model, p.client)
}

func (p *moonshotProvider) ProbeQuality(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: moonshotProviderID, ProbeType: "quality"}
}
func (p *moonshotProvider) ProbeEmbedding(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: moonshotProviderID, ProbeType: "embedding"}
}
func (p *moonshotProvider) ProbeStreaming(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: moonshotProviderID, ProbeType: "streaming"}
}
