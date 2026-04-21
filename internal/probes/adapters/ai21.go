package adapters

import (
	"context"
	"net/http"
	"time"

	"github.com/llmstatus/llmstatus/internal/httpclient"
	"github.com/llmstatus/llmstatus/internal/probes"
)

const (
	ai21DefaultBaseURL = "https://api.ai21.com/studio/v1"
	ai21ProviderID     = "ai21"
	ai21LightModel     = "jamba-mini"
	ai21LightProbeType = "light_inference"
	ai21ErrorDetailMax = 200
)

// AI21Option configures an AI21 Labs provider at construction time.
type AI21Option func(*ai21Provider)

// WithAI21BaseURL overrides the base URL. Intended for tests.
func WithAI21BaseURL(u string) AI21Option { return func(p *ai21Provider) { p.baseURL = u } }

// WithAI21HTTPClient overrides the HTTP client. Intended for tests.
func WithAI21HTTPClient(c *http.Client) AI21Option {
	return func(p *ai21Provider) { p.client = c }
}

// NewAI21Provider returns a probes.Provider backed by api.ai21.com.
func NewAI21Provider(apiKey, region string, opts ...AI21Option) probes.Provider {
	p := &ai21Provider{
		baseURL: ai21DefaultBaseURL,
		apiKey:  apiKey,
		region:  region,
		client:  httpclient.New(httpclient.Options{Timeout: 30 * time.Second}),
	}
	for _, o := range opts {
		o(p)
	}
	return p
}

type ai21Provider struct {
	baseURL, apiKey, region string
	client                  *http.Client
}

func (p *ai21Provider) ID() string       { return ai21ProviderID }
func (p *ai21Provider) Models() []string { return []string{ai21LightModel} }

func (p *ai21Provider) ProbeLightInference(ctx context.Context, model string) (probes.ProbeResult, error) {
	return probeOpenAICompat(ctx, p.baseURL, p.apiKey, p.region, ai21ProviderID, ai21LightProbeType, ai21ErrorDetailMax, model, p.client)
}

func (p *ai21Provider) ProbeQuality(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: ai21ProviderID, ProbeType: "quality"}
}
func (p *ai21Provider) ProbeEmbedding(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: ai21ProviderID, ProbeType: "embedding"}
}
func (p *ai21Provider) ProbeStreaming(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: ai21ProviderID, ProbeType: "streaming"}
}
