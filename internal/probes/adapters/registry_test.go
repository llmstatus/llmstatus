package adapters

import (
	"context"
	"testing"

	"github.com/llmstatus/llmstatus/internal/probes"
)

type stubProvider struct{ id string }

func (s *stubProvider) ID() string       { return s.id }
func (s *stubProvider) Models() []string { return []string{"m"} }
func (s *stubProvider) ProbeLightInference(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, nil
}
func (s *stubProvider) ProbeQuality(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, nil
}
func (s *stubProvider) ProbeEmbedding(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, nil
}
func (s *stubProvider) ProbeStreaming(_ context.Context, _ string) (probes.ProbeResult, error) {
	return probes.ProbeResult{}, nil
}

func TestRegistry_RegisterGetAll(t *testing.T) {
	// Isolate from any globals other tests may have registered.
	restore := registry
	registry = make(map[string]probes.Provider)
	t.Cleanup(func() { registry = restore })

	a := &stubProvider{id: "a"}
	b := &stubProvider{id: "b"}
	Register(a)
	Register(b)

	if got, ok := Get("a"); !ok || got.ID() != "a" {
		t.Errorf("Get(a): got (%v, %v)", got, ok)
	}
	if _, ok := Get("missing"); ok {
		t.Error(`Get("missing") should report !ok`)
	}

	all := All()
	if len(all) != 2 {
		t.Fatalf("All: got %d providers, want 2", len(all))
	}
	if all[0].ID() != "a" || all[1].ID() != "b" {
		t.Errorf("All: not sorted; got %q, %q", all[0].ID(), all[1].ID())
	}
}

func TestRegistry_RegisterOverwrites(t *testing.T) {
	restore := registry
	registry = make(map[string]probes.Provider)
	t.Cleanup(func() { registry = restore })

	Register(&stubProvider{id: "x"})
	Register(&stubProvider{id: "x"})

	if all := All(); len(all) != 1 {
		t.Errorf("All after overwrite: len=%d, want 1", len(all))
	}
}
