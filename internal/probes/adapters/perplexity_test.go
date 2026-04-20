package adapters

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/llmstatus/llmstatus/internal/probes"
)

func TestPerplexity_Identity(t *testing.T) {
	p := NewPerplexityProvider("sk-fake", "node-1")
	if p.ID() != "perplexity" {
		t.Errorf("ID: got %q", p.ID())
	}
}

func TestPerplexity_ProbeLightInference(t *testing.T) {
	testOpenAICompatAdapter(t, "perplexity", "perplexity",
		func(base string) probes.Provider {
			return NewPerplexityProvider("k", "r", WithPerplexityBaseURL(base))
		},
		perplexityLightModel,
	)
}

func TestPerplexity_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	p := NewPerplexityProvider("k", "r", WithPerplexityBaseURL(srv.URL), WithPerplexityHTTPClient(&http.Client{Timeout: 30 * time.Millisecond}))
	r, err := p.ProbeLightInference(context.Background(), perplexityLightModel)
	if err != nil || r.ErrorClass != probes.ErrorClassTimeout {
		t.Errorf("want timeout; err=%v class=%q", err, r.ErrorClass)
	}
}

func TestPerplexity_UnsupportedProbes(t *testing.T) {
	p := NewPerplexityProvider("k", "r")
	var ns *probes.ErrNotSupported
	for _, fn := range []func() error{
		func() error { _, e := p.ProbeQuality(context.Background(), "m"); return e },
		func() error { _, e := p.ProbeEmbedding(context.Background(), "m"); return e },
		func() error { _, e := p.ProbeStreaming(context.Background(), "m"); return e },
	} {
		if !errors.As(fn(), &ns) {
			t.Error("want ErrNotSupported")
		}
	}
}
