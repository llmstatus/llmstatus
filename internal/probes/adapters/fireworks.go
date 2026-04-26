package adapters

import (
	"net/http"

	"github.com/llmstatus/llmstatus/internal/probes"
)

const (
	fireworksDefaultBaseURL = "https://api.fireworks.ai/inference/v1"
	fireworksProviderID     = "fireworks"
	fireworksLightModel     = "accounts/fireworks/models/llama-v3p1-8b-instruct"
	fireworksLightProbeType = "light_inference"
	fireworksErrorDetailMax = 200
)

// FireworksOption configures a Fireworks AI provider at construction time.
type FireworksOption = compatOption

// WithNewFireworksProviderBaseURL overrides the base URL. Intended for tests.
func WithNewFireworksProviderBaseURL(u string) FireworksOption { return compatBaseURL(u) }

// WithNewFireworksProviderHTTPClient overrides the HTTP client. Intended for tests.
func WithNewFireworksProviderHTTPClient(c *http.Client) FireworksOption { return compatHTTPClient(c) }

// NewFireworksProvider returns a probes.Provider backed by api.fireworks.ai.
func NewFireworksProvider(apiKey, region string, opts ...FireworksOption) probes.Provider {
	return newOpenAICompatProvider(fireworksProviderID, fireworksDefaultBaseURL, apiKey, region, fireworksLightModel, fireworksLightProbeType, fireworksErrorDetailMax, opts...)
}
