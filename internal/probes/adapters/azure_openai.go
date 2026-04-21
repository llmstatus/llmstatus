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
	azureOpenAIDefaultBaseURL   = "https://%s.openai.azure.com"
	azureOpenAIDefaultAPIVersion = "2024-10-21"
	azureOpenAIProviderID       = "azure_openai"
	azureOpenAILightProbeType   = "light_inference"
	azureOpenAIErrorDetailMax   = 200
)

// AzureOpenAIOption configures an Azure OpenAI provider at construction time.
type AzureOpenAIOption func(*azureOpenAIProvider)

// WithAzureOpenAIBaseURL overrides the base URL (replaces https://{resource}.openai.azure.com).
// Intended for tests.
func WithAzureOpenAIBaseURL(u string) AzureOpenAIOption {
	return func(p *azureOpenAIProvider) { p.baseURL = u }
}

// WithAzureOpenAIHTTPClient overrides the HTTP client. Intended for tests.
func WithAzureOpenAIHTTPClient(c *http.Client) AzureOpenAIOption {
	return func(p *azureOpenAIProvider) { p.client = c }
}

// NewAzureOpenAIProvider returns a probes.Provider backed by Azure OpenAI.
// resource is the Azure resource name (e.g. "my-resource").
// deployment is the model deployment name (e.g. "gpt-4o-mini").
// apiVersion is the Azure API version (e.g. "2024-10-21").
func NewAzureOpenAIProvider(apiKey, resource, deployment, apiVersion, region string, opts ...AzureOpenAIOption) probes.Provider {
	p := &azureOpenAIProvider{
		baseURL:    fmt.Sprintf(azureOpenAIDefaultBaseURL, resource),
		apiKey:     apiKey,
		deployment: deployment,
		apiVersion: apiVersion,
		region:     region,
		client:     httpclient.New(httpclient.Options{Timeout: 30 * time.Second}),
	}
	for _, o := range opts {
		o(p)
	}
	return p
}

type azureOpenAIProvider struct {
	baseURL, apiKey, deployment, apiVersion, region string
	client                                          *http.Client
}

func (p *azureOpenAIProvider) ID() string       { return azureOpenAIProviderID }
func (p *azureOpenAIProvider) Models() []string { return []string{p.deployment} }

func (p *azureOpenAIProvider) ProbeLightInference(ctx context.Context, model string) (probes.ProbeResult, error) {
	return probeAzureOpenAI(ctx, p.baseURL, p.apiKey, p.deployment, p.apiVersion, p.region, azureOpenAIErrorDetailMax, p.client)
}

func (p *azureOpenAIProvider) ProbeQuality(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: azureOpenAIProviderID, ProbeType: "quality"}
}
func (p *azureOpenAIProvider) ProbeEmbedding(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: azureOpenAIProviderID, ProbeType: "embedding"}
}
func (p *azureOpenAIProvider) ProbeStreaming(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: azureOpenAIProviderID, ProbeType: "streaming"}
}

// probeAzureOpenAI fires a minimal chat completion against an Azure OpenAI deployment.
// Azure uses api-key header auth and a deployment-scoped URL path.
func probeAzureOpenAI(ctx context.Context, baseURL, apiKey, deployment, apiVersion, region string, maxDetail int, client *http.Client) (probes.ProbeResult, error) {
	started := time.Now()
	r := probes.ProbeResult{
		ProviderID: azureOpenAIProviderID,
		Model:      deployment,
		ProbeType:  azureOpenAILightProbeType,
		StartedAt:  started.UTC(),
		RegionID:   region,
	}

	payload, err := json.Marshal(openaiChatRequest{
		Model:     deployment,
		Messages:  []openaiChatMessage{{Role: "user", Content: "ping"}},
		MaxTokens: 1,
	})
	if err != nil {
		r.DurationMs = time.Since(started).Milliseconds()
		r.ErrorClass = probes.ErrorClassUnknown
		r.ErrorDetail = truncate(err.Error(), maxDetail)
		return r, err
	}

	url := fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=%s", baseURL, deployment, apiVersion)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		r.DurationMs = time.Since(started).Milliseconds()
		r.ErrorClass = probes.ErrorClassUnknown
		r.ErrorDetail = truncate(err.Error(), maxDetail)
		return r, err
	}
	// Azure uses api-key header, not Authorization: Bearer
	req.Header.Set("api-key", apiKey)
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
