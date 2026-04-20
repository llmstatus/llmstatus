//go:build livekeys

package adapters

import "os"

func init() {
	apiKey := os.Getenv("LLMS_MISTRAL_API_KEY")
	if apiKey == "" {
		return
	}
	Register(NewMistralProvider(apiKey, ""))
}
