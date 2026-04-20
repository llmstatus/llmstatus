//go:build livekeys

// This file wires the Anthropic adapter into the global registry only when
// the binary is built with `-tags livekeys`. See openai_register_livekeys.go
// for rationale.
//
// Build with: `go build -tags livekeys ./cmd/prober`

package adapters

import "os"

func init() {
	apiKey := os.Getenv("LLMS_ANTHROPIC_API_KEY")
	if apiKey == "" {
		return
	}
	Register(NewAnthropicProvider(apiKey, ""))
}
