package adapters

import (
	"net/http"

	"github.com/llmstatus/llmstatus/internal/probes"
)

const (
	perplexityDefaultBaseURL = "https://api.perplexity.ai"
	perplexityProviderID     = "perplexity"
	perplexityLightModel     = "sonar"
	perplexityLightProbeType = "light_inference"
	perplexityErrorDetailMax = 200
)

// PerplexityOption configures a Perplexity provider at construction time.
type PerplexityOption = compatOption

// WithPerplexityBaseURL overrides the base URL. Intended for tests.
func WithPerplexityBaseURL(u string) PerplexityOption { return compatBaseURL(u) }

// WithPerplexityHTTPClient overrides the HTTP client. Intended for tests.
func WithPerplexityHTTPClient(c *http.Client) PerplexityOption { return compatHTTPClient(c) }

// NewPerplexityProvider returns a probes.Provider backed by api.perplexity.ai.
func NewPerplexityProvider(apiKey, region string, opts ...PerplexityOption) probes.Provider {
	return newOpenAICompatProvider(perplexityProviderID, perplexityDefaultBaseURL, apiKey, region, perplexityLightModel, perplexityLightProbeType, perplexityErrorDetailMax, opts...)
}
