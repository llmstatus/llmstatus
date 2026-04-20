//go:build livekeys

package adapters

import "os"

func init() {
	apiKey := os.Getenv("LLMS_HUGGINGFACE_API_KEY")
	if apiKey == "" {
		return
	}
	Register(NewHuggingFaceProvider(apiKey, ""))
}
