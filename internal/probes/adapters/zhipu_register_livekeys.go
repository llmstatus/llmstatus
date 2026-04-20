//go:build livekeys

package adapters

import "os"

func init() {
	apiKey := os.Getenv("LLMS_ZHIPU_API_KEY")
	if apiKey == "" {
		return
	}
	Register(NewZhipuProvider(apiKey, ""))
}
