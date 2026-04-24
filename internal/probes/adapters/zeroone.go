package adapters

import (
	"context"
	"net/http"
	"time"

	"github.com/llmstatus/llmstatus/internal/httpclient"
	"github.com/llmstatus/llmstatus/internal/probes"
)

const (
	zerooneDefaultBaseURL = "https://api.01.ai/v1"
	zerooneProviderID     = "zeroone_ai"
	zerooneLightModel     = "yi-lightning"
	zerooneLightProbeType = "light_inference"
	zerooneErrorDetailMax = 200
)

// NewZeroOneProviderOption configures a NewZeroOneProvider provider at construction time.
type NewZeroOneProviderOption func(*zerooneProvider)

// WithNewZeroOneProviderBaseURL overrides the base URL. Intended for tests.
func WithNewZeroOneProviderBaseURL(u string) NewZeroOneProviderOption {
	return func(p *zerooneProvider) { p.baseURL = u }
}

// WithNewZeroOneProviderHTTPClient overrides the HTTP client. Intended for tests.
func WithNewZeroOneProviderHTTPClient(c *http.Client) NewZeroOneProviderOption {
	return func(p *zerooneProvider) { p.client = c }
}

// NewZeroOneProvider returns a probes.Provider backed by https://api.01.ai/v1.
func NewZeroOneProvider(apiKey, region string, opts ...NewZeroOneProviderOption) probes.Provider {
	p := &zerooneProvider{
		baseURL: zerooneDefaultBaseURL,
		apiKey:  apiKey,
		region:  region,
		client:  httpclient.New(httpclient.Options{Timeout: 30 * time.Second}),
	}
	for _, o := range opts {
		o(p)
	}
	return p
}

type zerooneProvider struct {
	baseURL, apiKey, region string
	client                  *http.Client
}

func (p *zerooneProvider) ID() string       { return zerooneProviderID }
func (p *zerooneProvider) Models() []string { return []string{zerooneLightModel} }

func (p *zerooneProvider) ProbeLightInference(ctx context.Context, model string) (probes.ProbeResult, error) {
	return probeOpenAICompat(ctx, p.baseURL, p.apiKey, p.region, zerooneProviderID, zerooneLightProbeType, zerooneErrorDetailMax, model, p.client)
}

func (p *zerooneProvider) ProbeQuality(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: zerooneProviderID, ProbeType: "quality"}
}
func (p *zerooneProvider) ProbeEmbedding(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: zerooneProviderID, ProbeType: "embedding"}
}
func (p *zerooneProvider) ProbeStreaming(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: zerooneProviderID, ProbeType: "streaming"}
}
