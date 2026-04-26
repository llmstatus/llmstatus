package adapters

import (
	"net/http"

	"github.com/llmstatus/llmstatus/internal/probes"
)

const (
	qwenDefaultBaseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	qwenProviderID     = "qwen"
	qwenLightModel     = "qwen-turbo"
	qwenLightProbeType = "light_inference"
	qwenErrorDetailMax = 200
)

// QwenOption configures a Qwen (Alibaba Cloud) provider at construction time.
type QwenOption = compatOption

// WithNewQwenProviderBaseURL overrides the base URL. Intended for tests.
func WithNewQwenProviderBaseURL(u string) QwenOption { return compatBaseURL(u) }

// WithNewQwenProviderHTTPClient overrides the HTTP client. Intended for tests.
func WithNewQwenProviderHTTPClient(c *http.Client) QwenOption { return compatHTTPClient(c) }

// NewQwenProvider returns a probes.Provider backed by dashscope.aliyuncs.com.
func NewQwenProvider(apiKey, region string, opts ...QwenOption) probes.Provider {
	return newOpenAICompatProvider(qwenProviderID, qwenDefaultBaseURL, apiKey, region, qwenLightModel, qwenLightProbeType, qwenErrorDetailMax, opts...)
}
