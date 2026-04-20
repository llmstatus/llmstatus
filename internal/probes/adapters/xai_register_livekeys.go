//go:build livekeys

package adapters

import "os"

func init() {
	apiKey := os.Getenv("LLMS_XAI_API_KEY")
	if apiKey == "" {
		return
	}
	Register(NewXAIProvider(apiKey, ""))
}
