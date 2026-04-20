package probes

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

const (
	defaultInterval     = 60 * time.Second
	defaultProbeTimeout = 30 * time.Second
	defaultConcurrency  = 8
)

// ResultSink receives probe results. Implementations include LogSink (default)
// and an HTTP sink (added in LLMS-004 when the ingest API is ready).
type ResultSink interface {
	Receive(ctx context.Context, r ProbeResult) error
}

// ProviderConfig pairs an adapter with the models it should probe.
// The caller (cmd/prober) builds this list from the DB and the adapter registry.
type ProviderConfig struct {
	Adapter Provider
	Models  []string
}

// Runner schedules ProbeLightInference for each Provider+Model pair at a
// fixed interval and emits results to a ResultSink.
type Runner struct {
	providers    []ProviderConfig
	sink         ResultSink
	regionID     string
	interval     time.Duration
	probeTimeout time.Duration
	concurrency  int
}

// Option configures a Runner.
type Option func(*Runner)

// WithInterval sets the time between probe rounds (default 60s).
func WithInterval(d time.Duration) Option {
	return func(r *Runner) { r.interval = d }
}

// WithProbeTimeout sets the per-probe context deadline (default 30s).
func WithProbeTimeout(d time.Duration) Option {
	return func(r *Runner) { r.probeTimeout = d }
}

// WithConcurrency sets the maximum number of simultaneous probe goroutines
// (default 8). Must be ≥ 1.
func WithConcurrency(n int) Option {
	return func(r *Runner) {
		if n < 1 {
			n = 1
		}
		r.concurrency = n
	}
}

// New creates a Runner. providers must not be empty. sink must not be nil.
func New(providers []ProviderConfig, sink ResultSink, regionID string, opts ...Option) *Runner {
	r := &Runner{
		providers:    providers,
		sink:         sink,
		regionID:     regionID,
		interval:     defaultInterval,
		probeTimeout: defaultProbeTimeout,
		concurrency:  defaultConcurrency,
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

// Run starts the probe loop. It fires one round immediately then repeats
// every interval. Blocks until ctx is cancelled; returns ctx.Err().
func (r *Runner) Run(ctx context.Context) error {
	r.runOnce(ctx)

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			r.runOnce(ctx)
		}
	}
}

// runOnce fires one probe per (provider, model) pair concurrently within
// the configured concurrency limit, then waits for all to complete.
func (r *Runner) runOnce(ctx context.Context) {
	sem := make(chan struct{}, r.concurrency)
	var wg sync.WaitGroup

	for i := range r.providers {
		pc := r.providers[i]
		for _, model := range pc.Models {
			model := model
			sem <- struct{}{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer func() { <-sem }()
				r.dispatchProbe(ctx, pc, model)
			}()
		}
	}
	wg.Wait()
}

// dispatchProbe runs all supported probe types for one (provider, model) pair
// sequentially and forwards each result to the sink. Unsupported probe types
// (ErrNotSupported) are silently skipped.
func (r *Runner) dispatchProbe(ctx context.Context, pc ProviderConfig, model string) {
	pCtx, cancel := context.WithTimeout(ctx, r.probeTimeout)
	defer cancel()

	probeFns := []func(context.Context, string) (ProbeResult, error){
		pc.Adapter.ProbeLightInference,
		pc.Adapter.ProbeQuality,
		pc.Adapter.ProbeEmbedding,
		pc.Adapter.ProbeStreaming,
	}

	for _, fn := range probeFns {
		result, err := fn(pCtx, model)
		if IsNotSupported(err) {
			continue
		}
		result.RegionID = r.regionID
		if sinkErr := r.sink.Receive(ctx, result); sinkErr != nil {
			slog.Error("runner: sink error",
				"provider", pc.Adapter.ID(),
				"model", model,
				"probe_type", result.ProbeType,
				"err", sinkErr,
			)
		}
	}
}

// LogSink writes every ProbeResult to the structured logger. It is the
// default sink used by cmd/prober when no ingest URL is configured.
type LogSink struct{}

// Receive logs the ProbeResult at Info level and returns nil.
func (LogSink) Receive(_ context.Context, r ProbeResult) error {
	slog.Info("probe",
		"provider", r.ProviderID,
		"model", r.Model,
		"probe_type", r.ProbeType,
		"success", r.Success,
		"duration_ms", r.DurationMs,
		"error_class", string(r.ErrorClass),
		"region", r.RegionID,
	)
	return nil
}
