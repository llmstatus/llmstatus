// Package probes defines the contract every AI-provider probe must satisfy.
//
// Each adapter under internal/probes/adapters implements Provider and is
// registered into the global adapter registry at process start.
package probes

import (
	"context"
	"time"
)

// Provider is implemented by one adapter per AI API provider.
//
// All methods take a context.Context as the first parameter and return a
// ProbeResult plus an error. Adapter authors should never return
// provider-specific error types from these methods; classify the error
// into ErrorClass before returning.
type Provider interface {
	// ID returns the stable, snake_case provider identifier
	// (e.g. "openai", "anthropic").
	ID() string

	// Models returns the list of model identifiers this adapter probes.
	Models() []string

	// ProbeLightInference issues a minimal inference request that exercises
	// the hot path without burning significant tokens.
	ProbeLightInference(ctx context.Context, model string) (ProbeResult, error)

	// ProbeQuality issues a representative inference whose response is
	// compared against a stored expected-output fixture.
	ProbeQuality(ctx context.Context, model string) (ProbeResult, error)

	// ProbeEmbedding issues a small embeddings request, if the provider
	// supports embeddings. Adapters without embeddings return
	// ErrNotSupported.
	ProbeEmbedding(ctx context.Context, model string) (ProbeResult, error)

	// ProbeStreaming issues a streaming inference and records the time to
	// first token. Adapters without streaming return ErrNotSupported.
	ProbeStreaming(ctx context.Context, model string) (ProbeResult, error)
}

// ErrorClass classifies provider errors into a fixed taxonomy defined in
// METHODOLOGY.md §5.4. Do not add new values without a methodology PR.
type ErrorClass string

// Error classification constants.
const (
	ErrorClassNone            ErrorClass = ""
	ErrorClassTimeout         ErrorClass = "timeout"
	ErrorClassRateLimit       ErrorClass = "rate_limit"
	ErrorClassAuth            ErrorClass = "auth"
	ErrorClassServer5xx       ErrorClass = "server_5xx"
	ErrorClassClient4xx       ErrorClass = "client_4xx"
	ErrorClassMalformedBody   ErrorClass = "malformed_body"
	ErrorClassQualityMismatch ErrorClass = "quality_mismatch"
	ErrorClassUnknown         ErrorClass = "unknown"
)

// ProbeResult is the single record emitted by every probe call. It is
// structured so it can be sent directly to the ingest service and stored
// as a time-series sample in InfluxDB.
type ProbeResult struct {
	ProviderID  string     `json:"provider_id"`
	Model       string     `json:"model"`
	ProbeType   string     `json:"probe_type"`
	StartedAt   time.Time  `json:"started_at"`
	DurationMs  int64      `json:"duration_ms"`
	HTTPStatus  int        `json:"http_status"`
	Success     bool       `json:"success"`
	ErrorClass  ErrorClass `json:"error_class,omitempty"`
	ErrorDetail string     `json:"error_detail,omitempty"`
	TokensIn    int        `json:"tokens_in,omitempty"`
	TokensOut   int        `json:"tokens_out,omitempty"`
	RegionID    string     `json:"region_id"`
	RawBodyHash string     `json:"raw_body_hash,omitempty"`
}

// ErrNotSupported is returned by Probe* methods when the provider does
// not support the requested probe type (e.g. an adapter with no embeddings).
type ErrNotSupported struct {
	ProviderID string
	ProbeType  string
}

func (e *ErrNotSupported) Error() string {
	return "probes: " + e.ProbeType + " not supported by provider " + e.ProviderID
}
