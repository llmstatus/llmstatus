package adapters

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/llmstatus/llmstatus/internal/probes"
)

// lightProbeConfig holds the provider-specific parameters for runLightProbe.
type lightProbeConfig struct {
	providerID   string
	probeType    string
	errorMax     int
	region       string
	buildRequest func(ctx context.Context, model string) (*http.Request, error)
	classifyResp func(r *probes.ProbeResult, status int, body []byte)
}

// runLightProbe executes the standard HTTP probe flow for custom-protocol
// adapters (those that cannot use probeOpenAICompat directly).
func runLightProbe(ctx context.Context, client *http.Client, model string, cfg lightProbeConfig) (probes.ProbeResult, error) {
	started := time.Now()
	r := probes.ProbeResult{
		ProviderID: cfg.providerID,
		Model:      model,
		ProbeType:  cfg.probeType,
		StartedAt:  started.UTC(),
		RegionID:   cfg.region,
	}

	req, err := cfg.buildRequest(ctx, model)
	if err != nil {
		r.DurationMs = time.Since(started).Milliseconds()
		r.ErrorClass = probes.ErrorClassUnknown
		r.ErrorDetail = truncate(err.Error(), cfg.errorMax)
		return r, err
	}

	resp, err := client.Do(req)
	r.DurationMs = time.Since(started).Milliseconds()
	if err != nil {
		r.ErrorClass = classifyNetError(err)
		r.ErrorDetail = truncate(err.Error(), cfg.errorMax)
		return r, nil
	}
	defer func() { _ = resp.Body.Close() }()

	r.HTTPStatus = resp.StatusCode
	body, _ := io.ReadAll(resp.Body)
	cfg.classifyResp(&r, resp.StatusCode, body)
	return r, nil
}
