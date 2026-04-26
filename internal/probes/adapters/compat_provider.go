package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/llmstatus/llmstatus/internal/httpclient"
	"github.com/llmstatus/llmstatus/internal/probes"
)

// compatOption configures an openAICompatProvider at construction time.
type compatOption func(*openAICompatProvider)

// openAICompatProvider is a generic probes.Provider for services that
// expose the OpenAI Chat Completions API and support only light inference.
type openAICompatProvider struct {
	providerID string
	baseURL    string
	apiKey     string
	region     string
	lightModel string
	probeType  string
	errorMax   int
	client     *http.Client
}

//nolint:unparam
func newOpenAICompatProvider(
	providerID, baseURL, apiKey, region, lightModel, probeType string,
	errorMax int,
	opts ...compatOption,
) probes.Provider {
	p := &openAICompatProvider{
		providerID: providerID,
		baseURL:    baseURL,
		apiKey:     apiKey,
		region:     region,
		lightModel: lightModel,
		probeType:  probeType,
		errorMax:   errorMax,
		client:     httpclient.New(httpclient.Options{Timeout: 30 * time.Second}),
	}
	for _, o := range opts {
		o(p)
	}
	return p
}

func compatBaseURL(u string) compatOption {
	return func(p *openAICompatProvider) { p.baseURL = u }
}

func compatHTTPClient(c *http.Client) compatOption {
	return func(p *openAICompatProvider) { p.client = c }
}

func (p *openAICompatProvider) ID() string       { return p.providerID }
func (p *openAICompatProvider) Models() []string { return []string{p.lightModel} }

func (p *openAICompatProvider) ProbeLightInference(ctx context.Context, model string) (probes.ProbeResult, error) {
	return probeOpenAICompat(ctx, p.baseURL, p.apiKey, p.region, p.providerID, p.probeType, p.errorMax, model, p.client)
}

func (p *openAICompatProvider) ProbeQuality(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: p.providerID, ProbeType: "quality"}
}

func (p *openAICompatProvider) ProbeEmbedding(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: p.providerID, ProbeType: "embedding"}
}

func (p *openAICompatProvider) ProbeStreaming(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: p.providerID, ProbeType: "streaming"}
}

// probeOpenAICompat is the shared probe implementation for all providers that
// expose the OpenAI Chat Completions API at /chat/completions.
//
//nolint:unparam
func probeOpenAICompat(ctx context.Context, baseURL, apiKey, region, providerID, probeType string, maxDetail int, model string, client *http.Client) (probes.ProbeResult, error) {
	started := time.Now()
	r := probes.ProbeResult{
		ProviderID: providerID,
		Model:      model,
		ProbeType:  probeType,
		StartedAt:  started.UTC(),
		RegionID:   region,
	}

	payload, err := json.Marshal(openaiChatRequest{
		Model:     model,
		Messages:  []openaiChatMessage{{Role: "user", Content: "ping"}},
		MaxTokens: 1,
	})
	if err != nil {
		r.DurationMs = time.Since(started).Milliseconds()
		r.ErrorClass = probes.ErrorClassUnknown
		r.ErrorDetail = truncate(err.Error(), maxDetail)
		return r, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		r.DurationMs = time.Since(started).Milliseconds()
		r.ErrorClass = probes.ErrorClassUnknown
		r.ErrorDetail = truncate(err.Error(), maxDetail)
		return r, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	r.DurationMs = time.Since(started).Milliseconds()
	if err != nil {
		r.ErrorClass = classifyNetError(err)
		r.ErrorDetail = truncate(err.Error(), maxDetail)
		return r, nil
	}
	defer func() { _ = resp.Body.Close() }()

	r.HTTPStatus = resp.StatusCode
	body, _ := io.ReadAll(resp.Body)
	classifyOpenAIResponse(&r, resp.StatusCode, body)
	return r, nil
}
