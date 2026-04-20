//go:build livekeys

package adapters

import "os"

func init() {
	apiKey := os.Getenv("LLMS_COHERE_API_KEY")
	if apiKey == "" {
		return
	}
	Register(NewCohereProvider(apiKey, ""))
}
