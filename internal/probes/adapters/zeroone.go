package adapters

import (
	"net/http"

	"github.com/llmstatus/llmstatus/internal/probes"
)

const (
	zerooneDefaultBaseURL = "https://api.01.ai/v1"
	zerooneProviderID     = "zeroone_ai"
	zerooneLightModel     = "yi-lightning"
	zerooneLightProbeType = "light_inference"
	zerooneErrorDetailMax = 200
)

// ZeroOneOption configures a 01.AI provider at construction time.
type ZeroOneOption = compatOption

// WithNewZeroOneProviderBaseURL overrides the base URL. Intended for tests.
func WithNewZeroOneProviderBaseURL(u string) ZeroOneOption { return compatBaseURL(u) }

// WithNewZeroOneProviderHTTPClient overrides the HTTP client. Intended for tests.
func WithNewZeroOneProviderHTTPClient(c *http.Client) ZeroOneOption { return compatHTTPClient(c) }

// NewZeroOneProvider returns a probes.Provider backed by api.01.ai.
func NewZeroOneProvider(apiKey, region string, opts ...ZeroOneOption) probes.Provider {
	return newOpenAICompatProvider(zerooneProviderID, zerooneDefaultBaseURL, apiKey, region, zerooneLightModel, zerooneLightProbeType, zerooneErrorDetailMax, opts...)
}
