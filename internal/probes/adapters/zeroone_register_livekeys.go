//go:build livekeys

package adapters

import "os"

func init() {
	apiKey := os.Getenv("LLMS_ZEROONE_API_KEY")
	if apiKey == "" {
		return
	}
	Register(NewZeroOneProvider(apiKey, ""))
}
