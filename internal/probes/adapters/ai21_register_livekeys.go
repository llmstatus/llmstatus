//go:build livekeys

package adapters

import "os"

func init() {
	apiKey := os.Getenv("LLMS_AI21_API_KEY")
	if apiKey == "" {
		return
	}
	Register(NewAI21Provider(apiKey, ""))
}
