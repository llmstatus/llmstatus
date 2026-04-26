package adapters

import (
	"net/http"

	"github.com/llmstatus/llmstatus/internal/probes"
)

const (
	moonshotDefaultBaseURL = "https://api.moonshot.cn/v1"
	moonshotProviderID     = "moonshot"
	moonshotLightModel     = "moonshot-v1-8k"
	moonshotLightProbeType = "light_inference"
	moonshotErrorDetailMax = 200
)

// NewMoonshotProviderOption configures a Moonshot provider at construction time.
type NewMoonshotProviderOption = compatOption

// WithNewMoonshotProviderBaseURL overrides the base URL. Intended for tests.
func WithNewMoonshotProviderBaseURL(u string) NewMoonshotProviderOption { return compatBaseURL(u) }

// WithNewMoonshotProviderHTTPClient overrides the HTTP client. Intended for tests.
func WithNewMoonshotProviderHTTPClient(c *http.Client) NewMoonshotProviderOption {
	return compatHTTPClient(c)
}

// NewMoonshotProvider returns a probes.Provider backed by api.moonshot.cn.
func NewMoonshotProvider(apiKey, region string, opts ...NewMoonshotProviderOption) probes.Provider {
	return newOpenAICompatProvider(moonshotProviderID, moonshotDefaultBaseURL, apiKey, region, moonshotLightModel, moonshotLightProbeType, moonshotErrorDetailMax, opts...)
}
