package adapters

import (
	"context"
	"net/http"
	"time"

	"github.com/llmstatus/llmstatus/internal/httpclient"
	"github.com/llmstatus/llmstatus/internal/probes"
)

const (
	zhipuDefaultBaseURL = "https://open.bigmodel.cn/api/paas/v4"
	zhipuProviderID     = "zhipu"
	zhipuLightModel     = "glm-4-flash"
	zhipuLightProbeType = "light_inference"
	zhipuErrorDetailMax = 200
)

// NewZhipuProviderOption configures a NewZhipuProvider provider at construction time.
type NewZhipuProviderOption func(*zhipuProvider)

// WithNewZhipuProviderBaseURL overrides the base URL. Intended for tests.
func WithNewZhipuProviderBaseURL(u string) NewZhipuProviderOption {
	return func(p *zhipuProvider) { p.baseURL = u }
}

// WithNewZhipuProviderHTTPClient overrides the HTTP client. Intended for tests.
func WithNewZhipuProviderHTTPClient(c *http.Client) NewZhipuProviderOption {
	return func(p *zhipuProvider) { p.client = c }
}

// NewZhipuProvider returns a probes.Provider backed by https://open.bigmodel.cn/api/paas/v4.
func NewZhipuProvider(apiKey, region string, opts ...NewZhipuProviderOption) probes.Provider {
	p := &zhipuProvider{
		baseURL: zhipuDefaultBaseURL,
		apiKey:  apiKey,
		region:  region,
		client:  httpclient.New(httpclient.Options{Timeout: 30 * time.Second}),
	}
	for _, o := range opts {
		o(p)
	}
	return p
}

type zhipuProvider struct {
	baseURL, apiKey, region string
	client                  *http.Client
}

func (p *zhipuProvider) ID() string       { return zhipuProviderID }
func (p *zhipuProvider) Models() []string { return []string{zhipuLightModel} }

func (p *zhipuProvider) ProbeLightInference(ctx context.Context, model string) (probes.ProbeResult, error) {
	return probeOpenAICompat(ctx, p.baseURL, p.apiKey, p.region, zhipuProviderID, zhipuLightProbeType, zhipuErrorDetailMax, model, p.client)
}

func (p *zhipuProvider) ProbeQuality(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: zhipuProviderID, ProbeType: "quality"}
}
func (p *zhipuProvider) ProbeEmbedding(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: zhipuProviderID, ProbeType: "embedding"}
}
func (p *zhipuProvider) ProbeStreaming(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: zhipuProviderID, ProbeType: "streaming"}
}
