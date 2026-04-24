package adapters

import (
	"context"
	"net/http"
	"time"

	"github.com/llmstatus/llmstatus/internal/httpclient"
	"github.com/llmstatus/llmstatus/internal/probes"
)

const (
	qwenDefaultBaseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	qwenProviderID     = "qwen"
	qwenLightModel     = "qwen-turbo"
	qwenLightProbeType = "light_inference"
	qwenErrorDetailMax = 200
)

// NewQwenProviderOption configures a NewQwenProvider provider at construction time.
type NewQwenProviderOption func(*qwenProvider)

// WithNewQwenProviderBaseURL overrides the base URL. Intended for tests.
func WithNewQwenProviderBaseURL(u string) NewQwenProviderOption {
	return func(p *qwenProvider) { p.baseURL = u }
}

// WithNewQwenProviderHTTPClient overrides the HTTP client. Intended for tests.
func WithNewQwenProviderHTTPClient(c *http.Client) NewQwenProviderOption {
	return func(p *qwenProvider) { p.client = c }
}

// NewQwenProvider returns a probes.Provider backed by https://dashscope.aliyuncs.com/compatible-mode/v1.
func NewQwenProvider(apiKey, region string, opts ...NewQwenProviderOption) probes.Provider {
	p := &qwenProvider{
		baseURL: qwenDefaultBaseURL,
		apiKey:  apiKey,
		region:  region,
		client:  httpclient.New(httpclient.Options{Timeout: 30 * time.Second}),
	}
	for _, o := range opts {
		o(p)
	}
	return p
}

type qwenProvider struct {
	baseURL, apiKey, region string
	client                  *http.Client
}

func (p *qwenProvider) ID() string       { return qwenProviderID }
func (p *qwenProvider) Models() []string { return []string{qwenLightModel} }

func (p *qwenProvider) ProbeLightInference(ctx context.Context, model string) (probes.ProbeResult, error) {
	return probeOpenAICompat(ctx, p.baseURL, p.apiKey, p.region, qwenProviderID, qwenLightProbeType, qwenErrorDetailMax, model, p.client)
}

func (p *qwenProvider) ProbeQuality(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: qwenProviderID, ProbeType: "quality"}
}
func (p *qwenProvider) ProbeEmbedding(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: qwenProviderID, ProbeType: "embedding"}
}
func (p *qwenProvider) ProbeStreaming(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: qwenProviderID, ProbeType: "streaming"}
}
