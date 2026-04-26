package adapters

import (
	"net/http"

	"github.com/llmstatus/llmstatus/internal/probes"
)

const (
	zhipuDefaultBaseURL = "https://open.bigmodel.cn/api/paas/v4"
	zhipuProviderID     = "zhipu"
	zhipuLightModel     = "glm-4-flash"
	zhipuLightProbeType = "light_inference"
	zhipuErrorDetailMax = 200
)

// ZhipuOption configures a Zhipu AI provider at construction time.
type ZhipuOption = compatOption

// WithNewZhipuProviderBaseURL overrides the base URL. Intended for tests.
func WithNewZhipuProviderBaseURL(u string) ZhipuOption { return compatBaseURL(u) }

// WithNewZhipuProviderHTTPClient overrides the HTTP client. Intended for tests.
func WithNewZhipuProviderHTTPClient(c *http.Client) ZhipuOption { return compatHTTPClient(c) }

// NewZhipuProvider returns a probes.Provider backed by open.bigmodel.cn.
func NewZhipuProvider(apiKey, region string, opts ...ZhipuOption) probes.Provider {
	return newOpenAICompatProvider(zhipuProviderID, zhipuDefaultBaseURL, apiKey, region, zhipuLightModel, zhipuLightProbeType, zhipuErrorDetailMax, opts...)
}
