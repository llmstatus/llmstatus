//go:build livekeys

package adapters

import "os"

func init() {
	apiKey := os.Getenv("LLMS_QWEN_API_KEY")
	if apiKey == "" {
		return
	}
	Register(NewQwenProvider(apiKey, ""))
}
