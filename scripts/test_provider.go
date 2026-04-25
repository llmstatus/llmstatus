//go:build ignore

// Quick smoke-test for a single provider adapter.
// Usage: go run ./scripts/test_provider.go <provider_id>
// Provider env vars must be set (same names as in docker-compose.yml).
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/llmstatus/llmstatus/internal/probes"
	"github.com/llmstatus/llmstatus/internal/probes/adapters"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run ./scripts/test_provider.go <provider_id>")
		os.Exit(1)
	}
	providerID := os.Args[1]

	provider, err := buildProvider(providerID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Printf("provider: %s  models: %v\n\n", provider.ID(), provider.Models())

	for _, model := range provider.Models() {
		for _, pt := range []string{"light_inference", "quality", "streaming", "embedding"} {
			result, runErr := runProbe(ctx, provider, model, pt)
			if runErr != nil {
				fmt.Printf("  [%-16s] %-30s  SKIP (%v)\n", pt, model, runErr)
				continue
			}
			status := "OK "
			if !result.Success {
				status = "FAIL"
			}
			fmt.Printf("  [%-16s] %-30s  %s  %dms  http=%d  err=%s\n",
				pt, model, status, result.DurationMs, result.HTTPStatus, result.ErrorClass)
			if result.ErrorDetail != "" {
				fmt.Printf("                                                    detail: %s\n", result.ErrorDetail)
			}
		}
		fmt.Println()
	}
}

func runProbe(ctx context.Context, p probes.Provider, model, probeType string) (probes.ProbeResult, error) {
	switch probeType {
	case "light_inference":
		r, err := p.ProbeLightInference(ctx, model)
		return r, skipIfNotSupported(err)
	case "quality":
		r, err := p.ProbeQuality(ctx, model)
		return r, skipIfNotSupported(err)
	case "streaming":
		r, err := p.ProbeStreaming(ctx, model)
		return r, skipIfNotSupported(err)
	case "embedding":
		r, err := p.ProbeEmbedding(ctx, model)
		return r, skipIfNotSupported(err)
	default:
		return probes.ProbeResult{}, fmt.Errorf("unknown probe type %q", probeType)
	}
}

func skipIfNotSupported(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := err.(*probes.ErrNotSupported); ok {
		return err
	}
	return nil
}

func buildProvider(id string) (probes.Provider, error) {
	env := func(k string) string { return os.Getenv(k) }
	switch id {
	case "openai":
		key := env("LLMS_OPENAI_API_KEY")
		if key == "" {
			return nil, fmt.Errorf("LLMS_OPENAI_API_KEY not set")
		}
		return adapters.NewOpenAIProvider(key, "local-test"), nil
	case "anthropic":
		key := env("LLMS_ANTHROPIC_API_KEY")
		if key == "" {
			return nil, fmt.Errorf("LLMS_ANTHROPIC_API_KEY not set")
		}
		return adapters.NewAnthropicProvider(key, "local-test"), nil
	case "google_gemini":
		key := env("LLMS_GEMINI_API_KEY")
		if key == "" {
			return nil, fmt.Errorf("LLMS_GEMINI_API_KEY not set")
		}
		return adapters.NewGeminiProvider(key, "local-test"), nil
	case "azure_openai":
		key := env("LLMS_AZURE_OPENAI_API_KEY")
		resource := env("LLMS_AZURE_OPENAI_RESOURCE")
		deployment := env("LLMS_AZURE_OPENAI_DEPLOYMENT")
		apiVersion := env("LLMS_AZURE_OPENAI_API_VERSION")
		if apiVersion == "" {
			apiVersion = "2024-10-21"
		}
		if key == "" || resource == "" || deployment == "" {
			return nil, fmt.Errorf("LLMS_AZURE_OPENAI_API_KEY / RESOURCE / DEPLOYMENT not set")
		}
		return adapters.NewAzureOpenAIProvider(key, resource, deployment, apiVersion, "local-test"), nil
	default:
		return nil, fmt.Errorf("unknown provider %q (add it to scripts/test_provider.go)", id)
	}
}

// keep the json import used
var _ = json.Marshal
