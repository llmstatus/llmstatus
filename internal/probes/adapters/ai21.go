package adapters

import (
	"net/http"

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
type AI21Option = compatOption

// WithAI21BaseURL overrides the base URL. Intended for tests.
func WithAI21BaseURL(u string) AI21Option { return compatBaseURL(u) }

// WithAI21HTTPClient overrides the HTTP client. Intended for tests.
func WithAI21HTTPClient(c *http.Client) AI21Option { return compatHTTPClient(c) }

// NewAI21Provider returns a probes.Provider backed by api.ai21.com.
func NewAI21Provider(apiKey, region string, opts ...AI21Option) probes.Provider {
	return newOpenAICompatProvider(ai21ProviderID, ai21DefaultBaseURL, apiKey, region, ai21LightModel, ai21LightProbeType, ai21ErrorDetailMax, opts...)
}
