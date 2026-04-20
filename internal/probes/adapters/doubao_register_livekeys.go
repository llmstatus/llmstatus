//go:build livekeys

package adapters

import "os"

func init() {
	apiKey := os.Getenv("LLMS_DOUBAO_API_KEY")
	if apiKey == "" {
		return
	}
	Register(NewDoubaoProvider(apiKey, ""))
}
