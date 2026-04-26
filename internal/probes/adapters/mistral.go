package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/llmstatus/llmstatus/internal/httpclient"
	"github.com/llmstatus/llmstatus/internal/probes"
)

const (
	mistralDefaultBaseURL = "https://api.mistral.ai/v1"
	mistralProviderID     = "mistral"
	mistralLightModel     = "mistral-small-latest"
	mistralLightProbeType = "light_inference"
	mistralErrorDetailMax = 200
)

// MistralOption configures a Mistral provider at construction time.
type MistralOption func(*mistralProvider)

// WithMistralBaseURL overrides the base URL. Intended for tests.
func WithMistralBaseURL(u string) MistralOption {
	return func(p *mistralProvider) { p.baseURL = u }
}

// WithMistralHTTPClient overrides the HTTP client. Intended for tests.
func WithMistralHTTPClient(c *http.Client) MistralOption {
	return func(p *mistralProvider) { p.client = c }
}

// NewMistralProvider returns a probes.Provider backed by api.mistral.ai.
// Mistral exposes an OpenAI-compatible Chat Completions API.
func NewMistralProvider(apiKey, region string, opts ...MistralOption) probes.Provider {
	p := &mistralProvider{
		baseURL: mistralDefaultBaseURL,
		apiKey:  apiKey,
		region:  region,
		client:  httpclient.New(httpclient.Options{Timeout: 30 * time.Second}),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

type mistralProvider struct {
	baseURL string
	apiKey  string
	region  string
	client  *http.Client
}

func (p *mistralProvider) ID() string       { return mistralProviderID }
func (p *mistralProvider) Models() []string { return []string{mistralLightModel} }

func (p *mistralProvider) ProbeLightInference(ctx context.Context, model string) (probes.ProbeResult, error) {
	return runLightProbe(ctx, p.client, model, lightProbeConfig{
		providerID:   mistralProviderID,
		probeType:    mistralLightProbeType,
		errorMax:     mistralErrorDetailMax,
		region:       p.region,
		buildRequest: p.buildRequest,
		classifyResp: classifyMistralResponse,
	})
}

func (p *mistralProvider) ProbeQuality(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: mistralProviderID, ProbeType: "quality"}
}
func (p *mistralProvider) ProbeEmbedding(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: mistralProviderID, ProbeType: "embedding"}
}
func (p *mistralProvider) ProbeStreaming(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: mistralProviderID, ProbeType: "streaming"}
}

func (p *mistralProvider) buildRequest(ctx context.Context, model string) (*http.Request, error) {
	body, err := json.Marshal(openaiChatRequest{
		Model:     model,
		Messages:  []openaiChatMessage{{Role: "user", Content: "ping"}},
		MaxTokens: 1,
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// mistralErrorEnvelope covers Mistral's flat error format: {"message": "..."}.
type mistralErrorEnvelope struct {
	Message string `json:"message"`
}

func classifyMistralResponse(r *probes.ProbeResult, status int, body []byte) {
	if status >= 200 && status < 300 {
		// 2xx uses the standard OpenAI chat completion envelope.
		var cr openaiChatResponse
		if err := json.Unmarshal(body, &cr); err != nil {
			r.ErrorClass = probes.ErrorClassMalformedBody
			r.ErrorDetail = truncate(string(body), mistralErrorDetailMax)
			return
		}
		r.Success = true
		r.TokensIn = cr.Usage.PromptTokens
		r.TokensOut = cr.Usage.CompletionTokens
		return
	}
	r.ErrorClass = classifyOpenAIStatus(status)
	r.ErrorDetail = parseMistralError(body)
}

func parseMistralError(body []byte) string {
	var e mistralErrorEnvelope
	if err := json.Unmarshal(body, &e); err == nil && e.Message != "" {
		return truncate(e.Message, mistralErrorDetailMax)
	}
	return truncate(string(body), mistralErrorDetailMax)
}
