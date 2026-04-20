package adapters

import (
	"context"
	"net/http"
	"time"

	"github.com/llmstatus/llmstatus/internal/httpclient"
	"github.com/llmstatus/llmstatus/internal/probes"
)

const (
	huggingfaceDefaultBaseURL = "https://router.huggingface.co/v1"
	huggingfaceProviderID     = "huggingface"
	huggingfaceLightModel     = "meta-llama/Llama-3.2-3B-Instruct"
	huggingfaceLightProbeType = "light_inference"
	huggingfaceErrorDetailMax = 200
)

// NewHuggingFaceProviderOption configures a NewHuggingFaceProvider provider at construction time.
type NewHuggingFaceProviderOption func(*huggingfaceProvider)

// WithNewHuggingFaceProviderBaseURL overrides the base URL. Intended for tests.
func WithNewHuggingFaceProviderBaseURL(u string) NewHuggingFaceProviderOption { return func(p *huggingfaceProvider) { p.baseURL = u } }

// WithNewHuggingFaceProviderHTTPClient overrides the HTTP client. Intended for tests.
func WithNewHuggingFaceProviderHTTPClient(c *http.Client) NewHuggingFaceProviderOption {
	return func(p *huggingfaceProvider) { p.client = c }
}

// NewHuggingFaceProvider returns a probes.Provider backed by https://router.huggingface.co/v1.
func NewHuggingFaceProvider(apiKey, region string, opts ...NewHuggingFaceProviderOption) probes.Provider {
	p := &huggingfaceProvider{
		baseURL: huggingfaceDefaultBaseURL,
		apiKey:  apiKey,
		region:  region,
		client:  httpclient.New(httpclient.Options{Timeout: 30 * time.Second}),
	}
	for _, o := range opts {
		o(p)
	}
	return p
}

type huggingfaceProvider struct {
	baseURL, apiKey, region string
	client                  *http.Client
}

func (p *huggingfaceProvider) ID() string       { return huggingfaceProviderID }
func (p *huggingfaceProvider) Models() []string { return []string{huggingfaceLightModel} }

func (p *huggingfaceProvider) ProbeLightInference(ctx context.Context, model string) (probes.ProbeResult, error) {
	return probeOpenAICompat(ctx, p.baseURL, p.apiKey, p.region, huggingfaceProviderID, huggingfaceLightProbeType, huggingfaceErrorDetailMax, model, p.client)
}

func (p *huggingfaceProvider) ProbeQuality(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: huggingfaceProviderID, ProbeType: "quality"}
}
func (p *huggingfaceProvider) ProbeEmbedding(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: huggingfaceProviderID, ProbeType: "embedding"}
}
func (p *huggingfaceProvider) ProbeStreaming(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: huggingfaceProviderID, ProbeType: "streaming"}
}
