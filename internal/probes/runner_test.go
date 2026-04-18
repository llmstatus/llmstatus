package probes_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/goleak"

	"github.com/llmstatus/llmstatus/internal/probes"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

// ---- test doubles --------------------------------------------------------

type fakeAdapter struct {
	id     string
	models []string
	result probes.ProbeResult
	delay  time.Duration
	calls  atomic.Int64
}

func (a *fakeAdapter) ID() string       { return a.id }
func (a *fakeAdapter) Models() []string { return a.models }

func (a *fakeAdapter) ProbeLightInference(ctx context.Context, model string) (probes.ProbeResult, error) {
	a.calls.Add(1)
	if a.delay > 0 {
		select {
		case <-time.After(a.delay):
		case <-ctx.Done():
			return probes.ProbeResult{}, ctx.Err()
		}
	}
	r := a.result
	r.ProviderID = a.id
	r.Model = model
	return r, nil
}

func (a *fakeAdapter) ProbeQuality(_ context.Context, model string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: a.id, ProbeType: "quality"}
}
func (a *fakeAdapter) ProbeEmbedding(_ context.Context, model string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: a.id, ProbeType: "embedding"}
}
func (a *fakeAdapter) ProbeStreaming(_ context.Context, model string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, &probes.ErrNotSupported{ProviderID: a.id, ProbeType: "streaming"}
}

type collectSink struct {
	mu      sync.Mutex
	results []probes.ProbeResult
}

func (s *collectSink) Receive(_ context.Context, r probes.ProbeResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.results = append(s.results, r)
	return nil
}

func (s *collectSink) snapshot() []probes.ProbeResult {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]probes.ProbeResult, len(s.results))
	copy(out, s.results)
	return out
}

// ---- tests ---------------------------------------------------------------

func TestRunner_RunOnce_AllProbesDispatched(t *testing.T) {
	adapter := &fakeAdapter{
		id:     "test_provider",
		models: []string{"model-a", "model-b", "model-c"},
		result: probes.ProbeResult{Success: true, ProbeType: "light_inference"},
	}
	sink := &collectSink{}

	r := probes.New(
		[]probes.ProviderConfig{{Adapter: adapter, Models: adapter.Models()}},
		sink,
		"us-west-2",
		probes.WithInterval(10*time.Second),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		_ = r.Run(ctx)
	}()

	// Wait until all 3 models have been probed at least once.
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if int(adapter.calls.Load()) >= 3 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	cancel()

	got := sink.snapshot()
	if len(got) < 3 {
		t.Errorf("expected ≥3 results, got %d", len(got))
	}
	for _, res := range got {
		if res.RegionID != "us-west-2" {
			t.Errorf("RegionID: got %q, want %q", res.RegionID, "us-west-2")
		}
		if res.ProviderID != "test_provider" {
			t.Errorf("ProviderID: got %q, want %q", res.ProviderID, "test_provider")
		}
	}
}

func TestRunner_ContextCancellation_StopsCleanly(t *testing.T) {
	adapter := &fakeAdapter{
		id:     "slow_provider",
		models: []string{"model-a"},
		delay:  100 * time.Millisecond,
		result: probes.ProbeResult{Success: true},
	}
	sink := &collectSink{}

	r := probes.New(
		[]probes.ProviderConfig{{Adapter: adapter, Models: adapter.Models()}},
		sink,
		"eu-west-1",
		probes.WithInterval(500*time.Millisecond),
	)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- r.Run(ctx)
	}()

	time.Sleep(150 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != context.Canceled {
			t.Errorf("Run returned %v, want context.Canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after context cancellation")
	}
}

func TestRunner_BoundedConcurrency(t *testing.T) {
	const numModels = 20
	const maxConcurrency = 4

	inFlight := atomic.Int64{}
	maxSeen := atomic.Int64{}

	slowAdapter := &fakeAdapter{
		id:     "concurrent_provider",
		models: make([]string, numModels),
	}
	for i := range slowAdapter.models {
		slowAdapter.models[i] = "model"
	}

	// Override ProbeLightInference to track in-flight count.
	trackingAdapter := &trackingConcurrencyAdapter{
		fakeAdapter: slowAdapter,
		inFlight:    &inFlight,
		maxSeen:     &maxSeen,
		holdFor:     20 * time.Millisecond,
	}

	sink := &collectSink{}
	r := probes.New(
		[]probes.ProviderConfig{{Adapter: trackingAdapter, Models: slowAdapter.Models()}},
		sink,
		"test",
		probes.WithConcurrency(maxConcurrency),
		probes.WithInterval(10*time.Second),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() { _ = r.Run(ctx) }()

	time.Sleep(500 * time.Millisecond)
	cancel()

	if got := maxSeen.Load(); got > int64(maxConcurrency) {
		t.Errorf("max concurrent probes = %d, want ≤ %d", got, maxConcurrency)
	}
}

type trackingConcurrencyAdapter struct {
	*fakeAdapter
	inFlight *atomic.Int64
	maxSeen  *atomic.Int64
	holdFor  time.Duration
}

func (a *trackingConcurrencyAdapter) ProbeLightInference(ctx context.Context, model string) (probes.ProbeResult, error) {
	cur := a.inFlight.Add(1)
	defer a.inFlight.Add(-1)

	for {
		seen := a.maxSeen.Load()
		if cur <= seen || a.maxSeen.CompareAndSwap(seen, cur) {
			break
		}
	}

	select {
	case <-time.After(a.holdFor):
	case <-ctx.Done():
	}
	return probes.ProbeResult{Success: true, ProviderID: a.id, Model: model}, nil
}

func TestRunner_ProbeTimeout_Respected(t *testing.T) {
	// Adapter that takes longer than the probe timeout.
	adapter := &fakeAdapter{
		id:     "tardy_provider",
		models: []string{"slow-model"},
		delay:  500 * time.Millisecond,
		result: probes.ProbeResult{Success: true},
	}
	sink := &collectSink{}

	r := probes.New(
		[]probes.ProviderConfig{{Adapter: adapter, Models: adapter.Models()}},
		sink,
		"ap-northeast-1",
		probes.WithProbeTimeout(50*time.Millisecond), // shorter than adapter delay
		probes.WithInterval(10*time.Second),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() { _ = r.Run(ctx) }()

	// Give the first runOnce time to fire and time out.
	time.Sleep(300 * time.Millisecond)
	cancel()

	// The probe was called (even though it timed out).
	if adapter.calls.Load() == 0 {
		t.Error("expected probe to be called at least once")
	}
}

func TestLogSink_Receive_NoError(t *testing.T) {
	sink := probes.LogSink{}
	err := sink.Receive(context.Background(), probes.ProbeResult{
		ProviderID: "openai",
		Model:      "gpt-4o-mini",
		Success:    true,
		DurationMs: 312,
		RegionID:   "us-west-2",
	})
	if err != nil {
		t.Errorf("LogSink.Receive returned unexpected error: %v", err)
	}
}
