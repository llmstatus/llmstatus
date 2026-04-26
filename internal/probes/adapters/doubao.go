package adapters

import (
	"net/http"

	"github.com/llmstatus/llmstatus/internal/probes"
)

const (
	doubaoDefaultBaseURL = "https://ark.cn-beijing.volces.com/api/v3"
	doubaoProviderID     = "doubao"
	doubaoLightModel     = "doubao-lite-4k"
	doubaoLightProbeType = "light_inference"
	doubaoErrorDetailMax = 200
)

// NewDoubaoProviderOption configures a Doubao provider at construction time.
type NewDoubaoProviderOption = compatOption

// WithNewDoubaoProviderBaseURL overrides the base URL. Intended for tests.
func WithNewDoubaoProviderBaseURL(u string) NewDoubaoProviderOption { return compatBaseURL(u) }

// WithNewDoubaoProviderHTTPClient overrides the HTTP client. Intended for tests.
func WithNewDoubaoProviderHTTPClient(c *http.Client) NewDoubaoProviderOption {
	return compatHTTPClient(c)
}

// NewDoubaoProvider returns a probes.Provider backed by ark.cn-beijing.volces.com.
func NewDoubaoProvider(apiKey, region string, opts ...NewDoubaoProviderOption) probes.Provider {
	return newOpenAICompatProvider(doubaoProviderID, doubaoDefaultBaseURL, apiKey, region, doubaoLightModel, doubaoLightProbeType, doubaoErrorDetailMax, opts...)
}
