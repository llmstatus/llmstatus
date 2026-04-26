package adapters

import (
	"net/http"

	"github.com/llmstatus/llmstatus/internal/probes"
)

const (
	minimaxDefaultBaseURL = "https://api.minimax.chat/v1"
	minimaxProviderID     = "minimax"
	minimaxLightModel     = "MiniMax-Text-01"
	minimaxLightProbeType = "light_inference"
	minimaxErrorDetailMax = 200
)

// MinimaxOption configures a MiniMax provider at construction time.
type MinimaxOption = compatOption

// WithNewMinimaxProviderBaseURL overrides the base URL. Intended for tests.
func WithNewMinimaxProviderBaseURL(u string) MinimaxOption { return compatBaseURL(u) }

// WithNewMinimaxProviderHTTPClient overrides the HTTP client. Intended for tests.
func WithNewMinimaxProviderHTTPClient(c *http.Client) MinimaxOption { return compatHTTPClient(c) }

// NewMinimaxProvider returns a probes.Provider backed by api.minimax.chat.
func NewMinimaxProvider(apiKey, region string, opts ...MinimaxOption) probes.Provider {
	return newOpenAICompatProvider(minimaxProviderID, minimaxDefaultBaseURL, apiKey, region, minimaxLightModel, minimaxLightProbeType, minimaxErrorDetailMax, opts...)
}
