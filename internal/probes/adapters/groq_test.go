package adapters

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/llmstatus/llmstatus/internal/probes"
)

func mustReadFixtureDir(t *testing.T, dir, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, "testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s/%s: %v", dir, name, err)
	}
	return b
}

func testOpenAICompatAdapter(t *testing.T, providerID, fixtureDir string,
	newFn func(baseURL string) probes.Provider,
	lightModel string,
) {
	t.Helper()

	cases := []struct {
		name         string
		httpStatus   int
		fixture      string
		wantSuccess  bool
		wantErrClass probes.ErrorClass
	}{
		{"success", http.StatusOK, "chat_completions_200.json", true, probes.ErrorClassNone},
		{"auth", http.StatusUnauthorized, "chat_completions_401.json", false, probes.ErrorClassAuth},
		{"rate_limit", http.StatusTooManyRequests, "chat_completions_429.json", false, probes.ErrorClassRateLimit},
		{"server_5xx", http.StatusInternalServerError, "chat_completions_500.json", false, probes.ErrorClassServer5xx},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := mustReadFixtureDir(t, fixtureDir, tc.fixture)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/chat/completions" {
					t.Errorf("path: got %q, want /chat/completions", r.URL.Path)
				}
				w.WriteHeader(tc.httpStatus)
				_, _ = w.Write(body)
			}))
			t.Cleanup(srv.Close)

			p := newFn(srv.URL)
			r, err := p.ProbeLightInference(context.Background(), lightModel)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if r.ProviderID != providerID {
				t.Errorf("ProviderID: got %q, want %q", r.ProviderID, providerID)
			}
			if r.Success != tc.wantSuccess {
				t.Errorf("Success: got %v, want %v", r.Success, tc.wantSuccess)
			}
			if r.ErrorClass != tc.wantErrClass {
				t.Errorf("ErrorClass: got %q, want %q", r.ErrorClass, tc.wantErrClass)
			}
			if tc.wantSuccess && r.TokensIn == 0 {
				t.Error("expected non-zero tokens_in on success")
			}
		})
	}
}

func testOpenAICompatTimeout(t *testing.T, p probes.Provider, model string) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	_ = srv // timeout via client, not server close
}

// ---- Groq -------------------------------------------------------------------

func TestGroq_Identity(t *testing.T) {
	p := NewGroqProvider("sk-fake", "node-1")
	if p.ID() != "groq" {
		t.Errorf("ID: got %q", p.ID())
	}
	if len(p.Models()) == 0 || p.Models()[0] != groqLightModel {
		t.Errorf("Models: got %v", p.Models())
	}
}

func TestGroq_ProbeLightInference(t *testing.T) {
	testOpenAICompatAdapter(t, "groq", "groq",
		func(base string) probes.Provider { return NewGroqProvider("k", "r", WithGroqBaseURL(base)) },
		groqLightModel,
	)
}

func TestGroq_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	p := NewGroqProvider("k", "r", WithGroqBaseURL(srv.URL), WithGroqHTTPClient(&http.Client{Timeout: 30 * time.Millisecond}))
	r, err := p.ProbeLightInference(context.Background(), groqLightModel)
	if err != nil || r.ErrorClass != probes.ErrorClassTimeout {
		t.Errorf("want timeout; err=%v class=%q", err, r.ErrorClass)
	}
}

func TestGroq_UnsupportedProbes(t *testing.T) {
	p := NewGroqProvider("k", "r")
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
