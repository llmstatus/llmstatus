package adapters

import (
	"net/http"

	"github.com/llmstatus/llmstatus/internal/probes"
)

const (
	huggingfaceDefaultBaseURL = "https://api-inference.huggingface.co/v1"
	huggingfaceProviderID     = "huggingface"
	huggingfaceLightModel     = "meta-llama/Llama-3.2-1B-Instruct"
	huggingfaceLightProbeType = "light_inference"
	huggingfaceErrorDetailMax = 200
)

// HuggingFaceOption configures a Hugging Face provider at construction time.
type HuggingFaceOption = compatOption

// WithNewHuggingFaceProviderBaseURL overrides the base URL. Intended for tests.
func WithNewHuggingFaceProviderBaseURL(u string) HuggingFaceOption { return compatBaseURL(u) }

// WithNewHuggingFaceProviderHTTPClient overrides the HTTP client. Intended for tests.
func WithNewHuggingFaceProviderHTTPClient(c *http.Client) HuggingFaceOption {
	return compatHTTPClient(c)
}

// NewHuggingFaceProvider returns a probes.Provider backed by api-inference.huggingface.co.
func NewHuggingFaceProvider(apiKey, region string, opts ...HuggingFaceOption) probes.Provider {
	return newOpenAICompatProvider(huggingfaceProviderID, huggingfaceDefaultBaseURL, apiKey, region, huggingfaceLightModel, huggingfaceLightProbeType, huggingfaceErrorDetailMax, opts...)
}
