//go:build livekeys

package adapters

import "os"

func init() {
	apiKey := os.Getenv("LLMS_AZURE_OPENAI_API_KEY")
	resource := os.Getenv("LLMS_AZURE_OPENAI_RESOURCE")
	deployment := os.Getenv("LLMS_AZURE_OPENAI_DEPLOYMENT")
	apiVersion := os.Getenv("LLMS_AZURE_OPENAI_API_VERSION")
	if apiKey == "" || resource == "" || deployment == "" {
		return
	}
	if apiVersion == "" {
		apiVersion = azureOpenAIDefaultAPIVersion
	}
	Register(NewAzureOpenAIProvider(apiKey, resource, deployment, apiVersion, ""))
}
