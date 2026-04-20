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

const (
	cohereDefaultBaseURL = "https://api.cohere.com/v2"
	cohereProviderID     = "cohere"
	cohereLightModel     = "command-r"
	cohereLightProbeType = "light_inference"
	cohereErrorDetailMax = 200
)

// CohereOption configures a Cohere provider at construction time.
type CohereOption func(*cohereProvider)

// WithCohereBaseURL overrides the base URL. Intended for tests.
func WithCohereBaseURL(u string) CohereOption { return func(p *cohereProvider) { p.baseURL = u } }

// WithCohereHTTPClient overrides the HTTP client. Intended for tests.
func WithCohereHTTPClient(c *http.Client) CohereOption { return func(p *cohereProvider) { p.client = c } }

// NewCohereProvider returns a probes.Provider backed by api.cohere.com.
func NewCohereProvider(apiKey, region string, opts ...CohereOption) probes.Provider {
	p := &cohereProvider{
		baseURL: cohereDefaultBaseURL,
		apiKey:  apiKey,
		region:  region,
		client:  httpclient.New(httpclient.Options{Timeout: 30 * time.Second}),
	}
	for _, o := range opts {
		o(p)
	}
	return p
}

type cohereProvider struct {
	baseURL, apiKey, region string
	client                  *http.Client
}

func (p *cohereProvider) ID() string       { return cohereProviderID }
func (p *cohereProvider) Models() []string { return []string{cohereLightModel} }

func (p *cohereProvider) ProbeLightInference(ctx context.Context, model string) (probes.ProbeResult, error) {
	started := time.Now()
	r := probes.ProbeResult{
		ProviderID: cohereProviderID,
		Model:      model,
		ProbeType:  cohereLightProbeType,
		StartedAt:  started.UTC(),
		RegionID:   p.region,
	}

	req, err := p.buildRequest(ctx, model)
	if err != nil {
		r.DurationMs = time.Since(started).Milliseconds()
		r.ErrorClass = probes.ErrorClassUnknown
		r.ErrorDetail = truncate(err.Error(), cohereErrorDetailMax)
		return r, err
	}

	resp, err := p.client.Do(req)
	r.DurationMs = time.Since(started).Milliseconds()
	if err != nil {
		r.ErrorClass = classifyNetError(err)
		r.ErrorDetail = truncate(err.Error(), cohereErrorDetailMax)
		return r, nil
	}
	defer func() { _ = resp.Body.Close() }()

	r.HTTPStatus = resp.StatusCode
	body, _ := io.ReadAll(resp.Body)
	classifyCohereResponse(&r, resp.StatusCode, body)
	return r, nil
}

func (p *cohereProvider) ProbeQuality(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: cohereProviderID, ProbeType: "quality"}
}
func (p *cohereProvider) ProbeEmbedding(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: cohereProviderID, ProbeType: "embedding"}
}
func (p *cohereProvider) ProbeStreaming(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: cohereProviderID, ProbeType: "streaming"}
}

// ---- Cohere v2 API types ----------------------------------------------------

type cohereChatRequest struct {
	Model    string              `json:"model"`
	Messages []cohereChatMessage `json:"messages"`
}

type cohereChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type cohereChatResponse struct {
	Usage struct {
		BilledUnits struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"billed_units"`
	} `json:"usage"`
}

// cohereErrorEnvelope covers Cohere's flat {"message": "..."} error format.
type cohereErrorEnvelope struct {
	Message string `json:"message"`
}

func (p *cohereProvider) buildRequest(ctx context.Context, model string) (*http.Request, error) {
	body, err := json.Marshal(cohereChatRequest{
		Model:    model,
		Messages: []cohereChatMessage{{Role: "user", Content: "ping"}},
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func classifyCohereResponse(r *probes.ProbeResult, status int, body []byte) {
	if status >= 200 && status < 300 {
		var cr cohereChatResponse
		if err := json.Unmarshal(body, &cr); err != nil {
			r.ErrorClass = probes.ErrorClassMalformedBody
			r.ErrorDetail = truncate(string(body), cohereErrorDetailMax)
			return
		}
		r.Success = true
		r.TokensIn = cr.Usage.BilledUnits.InputTokens
		r.TokensOut = cr.Usage.BilledUnits.OutputTokens
		return
	}
	r.ErrorClass = classifyOpenAIStatus(status)
	r.ErrorDetail = parseCohereError(body)
}

func parseCohereError(body []byte) string {
	var e cohereErrorEnvelope
	if err := json.Unmarshal(body, &e); err == nil && e.Message != "" {
		return truncate(e.Message, cohereErrorDetailMax)
	}
	return truncate(string(body), cohereErrorDetailMax)
}
