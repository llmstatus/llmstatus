package adapters

import (
	"net/http"

	"github.com/llmstatus/llmstatus/internal/probes"
)

const (
	togetherDefaultBaseURL = "https://api.together.xyz/v1"
	togetherProviderID     = "together_ai"
	togetherLightModel     = "meta-llama/Meta-Llama-3.1-8B-Instruct-Turbo"
	togetherLightProbeType = "light_inference"
	togetherErrorDetailMax = 200
)

// TogetherOption configures a Together AI provider at construction time.
type TogetherOption = compatOption

// WithTogetherBaseURL overrides the base URL. Intended for tests.
func WithTogetherBaseURL(u string) TogetherOption { return compatBaseURL(u) }

// WithTogetherHTTPClient overrides the HTTP client. Intended for tests.
func WithTogetherHTTPClient(c *http.Client) TogetherOption { return compatHTTPClient(c) }

// NewTogetherProvider returns a probes.Provider backed by api.together.xyz.
func NewTogetherProvider(apiKey, region string, opts ...TogetherOption) probes.Provider {
	return newOpenAICompatProvider(togetherProviderID, togetherDefaultBaseURL, apiKey, region, togetherLightModel, togetherLightProbeType, togetherErrorDetailMax, opts...)
}
