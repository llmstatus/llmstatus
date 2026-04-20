//go:build livekeys

package adapters

import "os"

func init() {
	apiKey := os.Getenv("LLMS_DEEPSEEK_API_KEY")
	if apiKey == "" {
		return
	}
	Register(NewDeepSeekProvider(apiKey, ""))
}
