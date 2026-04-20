//go:build livekeys

package adapters

import "os"

func init() {
	apiKey := os.Getenv("LLMS_PERPLEXITY_API_KEY")
	if apiKey == "" {
		return
	}
	Register(NewPerplexityProvider(apiKey, ""))
}
