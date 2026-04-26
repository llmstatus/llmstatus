package adapters

import (
	"net/http"

	"github.com/llmstatus/llmstatus/internal/probes"
)

const (
	xaiDefaultBaseURL = "https://api.x.ai/v1"
	xaiProviderID     = "xai"
	xaiLightModel     = "grok-3-mini"
	xaiLightProbeType = "light_inference"
	xaiErrorDetailMax = 200
)

// XAIOption configures an xAI (Grok) provider at construction time.
type XAIOption = compatOption

// WithXAIBaseURL overrides the base URL. Intended for tests.
func WithXAIBaseURL(u string) XAIOption { return compatBaseURL(u) }

// WithXAIHTTPClient overrides the HTTP client. Intended for tests.
func WithXAIHTTPClient(c *http.Client) XAIOption { return compatHTTPClient(c) }

// NewXAIProvider returns a probes.Provider backed by api.x.ai.
func NewXAIProvider(apiKey, region string, opts ...XAIOption) probes.Provider {
	return newOpenAICompatProvider(xaiProviderID, xaiDefaultBaseURL, apiKey, region, xaiLightModel, xaiLightProbeType, xaiErrorDetailMax, opts...)
}
