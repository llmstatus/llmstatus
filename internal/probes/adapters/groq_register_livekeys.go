//go:build livekeys

package adapters

import "os"

func init() {
	apiKey := os.Getenv("LLMS_GROQ_API_KEY")
	if apiKey == "" {
		return
	}
	Register(NewGroqProvider(apiKey, ""))
}
