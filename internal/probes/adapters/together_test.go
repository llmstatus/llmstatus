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

func TestTogether_Identity(t *testing.T) {
	p := NewTogetherProvider("sk-fake", "node-1")
	if p.ID() != "together_ai" {
		t.Errorf("ID: got %q", p.ID())
	}
}

func TestTogether_ProbeLightInference(t *testing.T) {
	testOpenAICompatAdapter(t, "together_ai", "together",
		func(base string) probes.Provider {
			return NewTogetherProvider("k", "r", WithTogetherBaseURL(base))
		},
		togetherLightModel,
	)
}

func TestTogether_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	p := NewTogetherProvider("k", "r", WithTogetherBaseURL(srv.URL), WithTogetherHTTPClient(&http.Client{Timeout: 30 * time.Millisecond}))
	r, err := p.ProbeLightInference(context.Background(), togetherLightModel)
	if err != nil || r.ErrorClass != probes.ErrorClassTimeout {
		t.Errorf("want timeout; err=%v class=%q", err, r.ErrorClass)
	}
}

func TestTogether_UnsupportedProbes(t *testing.T) {
	p := NewTogetherProvider("k", "r")
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
