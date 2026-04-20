package adapters

import (
	"context"
	"net/http"
	"time"

	"github.com/llmstatus/llmstatus/internal/httpclient"
	"github.com/llmstatus/llmstatus/internal/probes"
)

const (
	doubaoDefaultBaseURL = "https://ark.cn-beijing.volces.com/api/v3"
	doubaoProviderID     = "doubao"
	doubaoLightModel     = "doubao-lite-4k"
	doubaoLightProbeType = "light_inference"
	doubaoErrorDetailMax = 200
)

// NewDoubaoProviderOption configures a NewDoubaoProvider provider at construction time.
type NewDoubaoProviderOption func(*doubaoProvider)

// WithNewDoubaoProviderBaseURL overrides the base URL. Intended for tests.
func WithNewDoubaoProviderBaseURL(u string) NewDoubaoProviderOption { return func(p *doubaoProvider) { p.baseURL = u } }

// WithNewDoubaoProviderHTTPClient overrides the HTTP client. Intended for tests.
func WithNewDoubaoProviderHTTPClient(c *http.Client) NewDoubaoProviderOption {
	return func(p *doubaoProvider) { p.client = c }
}

// NewDoubaoProvider returns a probes.Provider backed by https://ark.cn-beijing.volces.com/api/v3.
func NewDoubaoProvider(apiKey, region string, opts ...NewDoubaoProviderOption) probes.Provider {
	p := &doubaoProvider{
		baseURL: doubaoDefaultBaseURL,
		apiKey:  apiKey,
		region:  region,
		client:  httpclient.New(httpclient.Options{Timeout: 30 * time.Second}),
	}
	for _, o := range opts {
		o(p)
	}
	return p
}

type doubaoProvider struct {
	baseURL, apiKey, region string
	client                  *http.Client
}

func (p *doubaoProvider) ID() string       { return doubaoProviderID }
func (p *doubaoProvider) Models() []string { return []string{doubaoLightModel} }

func (p *doubaoProvider) ProbeLightInference(ctx context.Context, model string) (probes.ProbeResult, error) {
	return probeOpenAICompat(ctx, p.baseURL, p.apiKey, p.region, doubaoProviderID, doubaoLightProbeType, doubaoErrorDetailMax, model, p.client)
}

func (p *doubaoProvider) ProbeQuality(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: doubaoProviderID, ProbeType: "quality"}
}
func (p *doubaoProvider) ProbeEmbedding(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: doubaoProviderID, ProbeType: "embedding"}
}
func (p *doubaoProvider) ProbeStreaming(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: doubaoProviderID, ProbeType: "streaming"}
}
