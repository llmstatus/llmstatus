package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/llmstatus/llmstatus/internal/httpclient"
	"github.com/llmstatus/llmstatus/internal/probes"
)

const (
	geminiDefaultBaseURL = "https://generativelanguage.googleapis.com"
	geminiProviderID     = "google_gemini"
	geminiLightModel     = "gemini-2.0-flash"
	geminiLightProbeType = "light_inference"
	geminiErrorDetailMax = 200
)

// GeminiOption configures a Gemini provider at construction time.
type GeminiOption func(*geminiProvider)

// WithGeminiBaseURL overrides the base URL. Intended for tests.
func WithGeminiBaseURL(u string) GeminiOption {
	return func(p *geminiProvider) { p.baseURL = u }
}

// WithGeminiHTTPClient overrides the HTTP client. Intended for tests.
func WithGeminiHTTPClient(c *http.Client) GeminiOption {
	return func(p *geminiProvider) { p.client = c }
}

// NewGeminiProvider returns a probes.Provider backed by the Google Gemini API.
func NewGeminiProvider(apiKey, region string, opts ...GeminiOption) probes.Provider {
	p := &geminiProvider{
		baseURL: geminiDefaultBaseURL,
		apiKey:  apiKey,
		region:  region,
		client:  httpclient.New(httpclient.Options{Timeout: 30 * time.Second}),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

type geminiProvider struct {
	baseURL string
	apiKey  string
	region  string
	client  *http.Client
}

func (p *geminiProvider) ID() string       { return geminiProviderID }
func (p *geminiProvider) Models() []string { return []string{geminiLightModel} }

func (p *geminiProvider) ProbeLightInference(ctx context.Context, model string) (probes.ProbeResult, error) {
	started := time.Now()
	r := probes.ProbeResult{
		ProviderID: geminiProviderID,
		Model:      model,
		ProbeType:  geminiLightProbeType,
		StartedAt:  started.UTC(),
		RegionID:   p.region,
	}

	req, err := p.buildRequest(ctx, model)
	if err != nil {
		r.DurationMs = time.Since(started).Milliseconds()
		r.ErrorClass = probes.ErrorClassUnknown
		r.ErrorDetail = truncate(err.Error(), geminiErrorDetailMax)
		return r, err
	}

	resp, err := p.client.Do(req)
	r.DurationMs = time.Since(started).Milliseconds()
	if err != nil {
		r.ErrorClass = classifyNetError(err)
		r.ErrorDetail = truncate(err.Error(), geminiErrorDetailMax)
		return r, nil
	}
	defer func() { _ = resp.Body.Close() }()

	r.HTTPStatus = resp.StatusCode
	body, _ := io.ReadAll(resp.Body)
	classifyGeminiResponse(&r, resp.StatusCode, body)
	return r, nil
}

func (p *geminiProvider) ProbeQuality(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: geminiProviderID, ProbeType: "quality"}
}

func (p *geminiProvider) ProbeEmbedding(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: geminiProviderID, ProbeType: "embedding"}
}

func (p *geminiProvider) ProbeStreaming(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: geminiProviderID, ProbeType: "streaming"}
}

// gemini request/response types

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerateRequest struct {
	Contents         []geminiContent `json:"contents"`
	GenerationConfig geminiGenConfig `json:"generationConfig"`
}

type geminiGenConfig struct {
	MaxOutputTokens int `json:"maxOutputTokens"`
}

type geminiGenerateResponse struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
	} `json:"usageMetadata"`
}

type geminiErrorEnvelope struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

func (p *geminiProvider) buildRequest(ctx context.Context, model string) (*http.Request, error) {
	payload := geminiGenerateRequest{
		Contents:         []geminiContent{{Parts: []geminiPart{{Text: "ping"}}, Role: "user"}},
		GenerationConfig: geminiGenConfig{MaxOutputTokens: 1},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", p.baseURL, model, p.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func classifyGeminiResponse(r *probes.ProbeResult, status int, body []byte) {
	if status >= 200 && status < 300 {
		var cr geminiGenerateResponse
		if err := json.Unmarshal(body, &cr); err != nil || len(cr.Candidates) == 0 {
			r.ErrorClass = probes.ErrorClassMalformedBody
			r.ErrorDetail = truncate(string(body), geminiErrorDetailMax)
			return
		}
		r.Success = true
		r.TokensIn = cr.UsageMetadata.PromptTokenCount
		r.TokensOut = cr.UsageMetadata.CandidatesTokenCount
		return
	}
	r.ErrorClass = classifyGeminiStatus(status)
	r.ErrorDetail = parseGeminiError(body)
}

func classifyGeminiStatus(status int) probes.ErrorClass {
	switch {
	case status == http.StatusUnauthorized, status == http.StatusForbidden:
		return probes.ErrorClassAuth
	case status == http.StatusTooManyRequests:
		return probes.ErrorClassRateLimit
	case status >= 500:
		return probes.ErrorClassServer5xx
	case status >= 400:
		return probes.ErrorClassClient4xx
	default:
		return probes.ErrorClassUnknown
	}
}

func parseGeminiError(body []byte) string {
	var e geminiErrorEnvelope
	if err := json.Unmarshal(body, &e); err == nil && e.Error.Message != "" {
		return truncate(e.Error.Message, geminiErrorDetailMax)
	}
	return truncate(string(body), geminiErrorDetailMax)
}
