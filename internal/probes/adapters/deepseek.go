package adapters

import (
	"net/http"

	"github.com/llmstatus/llmstatus/internal/probes"
)

const (
	deepseekDefaultBaseURL = "https://api.deepseek.com/v1"
	deepseekProviderID     = "deepseek"
	deepseekLightModel     = "deepseek-chat"
	deepseekLightProbeType = "light_inference"
	deepseekErrorDetailMax = 200
)

// DeepSeekOption configures a DeepSeek provider at construction time.
type DeepSeekOption = compatOption

// WithDeepSeekBaseURL overrides the base URL. Intended for tests.
func WithDeepSeekBaseURL(u string) DeepSeekOption { return compatBaseURL(u) }

// WithDeepSeekHTTPClient overrides the HTTP client. Intended for tests.
func WithDeepSeekHTTPClient(c *http.Client) DeepSeekOption { return compatHTTPClient(c) }

// NewDeepSeekProvider returns a probes.Provider backed by api.deepseek.com.
// DeepSeek exposes an OpenAI-compatible Chat Completions API.
func NewDeepSeekProvider(apiKey, region string, opts ...DeepSeekOption) probes.Provider {
	return newOpenAICompatProvider(deepseekProviderID, deepseekDefaultBaseURL, apiKey, region, deepseekLightModel, deepseekLightProbeType, deepseekErrorDetailMax, opts...)
}
