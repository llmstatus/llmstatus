//go:build livekeys

// This file wires the OpenAI adapter into the global registry only when
// the binary is built with `-tags livekeys`. In dev and CI we want the
// adapter's code to compile and be testable, but never to fire live
// HTTP calls from the prober.
//
// Build with: `go build -tags livekeys ./cmd/prober`

package adapters

import "os"

func init() {
	apiKey := os.Getenv("LLMS_OPENAI_API_KEY")
	region := os.Getenv("LLMS_REGION_ID")
	if apiKey == "" || region == "" {
		return
	}
	Register(NewOpenAIProvider(apiKey, region))
}
