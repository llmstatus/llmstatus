package adapters

import (
	"net/http"

	"github.com/llmstatus/llmstatus/internal/probes"
)

const (
	groqDefaultBaseURL = "https://api.groq.com/openai/v1"
	groqProviderID     = "groq"
	groqLightModel     = "llama-3.3-70b-versatile"
	groqLightProbeType = "light_inference"
	groqErrorDetailMax = 200
)

// GroqOption configures a Groq provider at construction time.
type GroqOption = compatOption

// WithGroqBaseURL overrides the base URL. Intended for tests.
func WithGroqBaseURL(u string) GroqOption { return compatBaseURL(u) }

// WithGroqHTTPClient overrides the HTTP client. Intended for tests.
func WithGroqHTTPClient(c *http.Client) GroqOption { return compatHTTPClient(c) }

// NewGroqProvider returns a probes.Provider backed by api.groq.com.
func NewGroqProvider(apiKey, region string, opts ...GroqOption) probes.Provider {
	return newOpenAICompatProvider(groqProviderID, groqDefaultBaseURL, apiKey, region, groqLightModel, groqLightProbeType, groqErrorDetailMax, opts...)
}
