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

func TestXAI_Identity(t *testing.T) {
	p := NewXAIProvider("sk-fake", "node-1")
	if p.ID() != "xai" {
		t.Errorf("ID: got %q", p.ID())
	}
}

func TestXAI_ProbeLightInference(t *testing.T) {
	testOpenAICompatAdapter(t, "xai", "xai",
		func(base string) probes.Provider { return NewXAIProvider("k", "r", WithXAIBaseURL(base)) },
		xaiLightModel,
	)
}

func TestXAI_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	p := NewXAIProvider("k", "r", WithXAIBaseURL(srv.URL), WithXAIHTTPClient(&http.Client{Timeout: 30 * time.Millisecond}))
	r, err := p.ProbeLightInference(context.Background(), xaiLightModel)
	if err != nil || r.ErrorClass != probes.ErrorClassTimeout {
		t.Errorf("want timeout; err=%v class=%q", err, r.ErrorClass)
	}
}

func TestXAI_UnsupportedProbes(t *testing.T) {
	p := NewXAIProvider("k", "r")
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
