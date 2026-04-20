package adapters

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/llmstatus/llmstatus/internal/probes"
)

func mustReadMistralFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("mistral", "testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return b
}

func TestMistral_Identity(t *testing.T) {
	p := NewMistralProvider("fake-key", "node-1")
	if got := p.ID(); got != "mistral" {
		t.Errorf("ID: got %q, want mistral", got)
	}
	if got := p.Models(); len(got) == 0 || got[0] != "mistral-small-latest" {
		t.Errorf("Models: got %v", got)
	}
}

func TestMistral_ProbeLightInference(t *testing.T) {
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
			body := mustReadMistralFixture(t, tc.fixture)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/chat/completions" {
					t.Errorf("path: got %q, want /chat/completions", r.URL.Path)
				}
				w.WriteHeader(tc.httpStatus)
				_, _ = w.Write(body)
			}))
			t.Cleanup(srv.Close)

			p := NewMistralProvider("fake-key", "node-1", WithMistralBaseURL(srv.URL))
			r, err := p.ProbeLightInference(context.Background(), "mistral-small-latest")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if r.ProviderID != "mistral" {
				t.Errorf("ProviderID: got %q, want mistral", r.ProviderID)
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

func TestMistral_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	p := NewMistralProvider("fake-key", "node-1",
		WithMistralBaseURL(srv.URL),
		WithMistralHTTPClient(&http.Client{Timeout: 30 * time.Millisecond}),
	)
	r, err := p.ProbeLightInference(context.Background(), "mistral-small-latest")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.ErrorClass != probes.ErrorClassTimeout {
		t.Errorf("ErrorClass: got %q, want timeout", r.ErrorClass)
	}
}

func TestMistral_UnsupportedProbes(t *testing.T) {
	p := NewMistralProvider("fake-key", "node-1")
	var ns *probes.ErrNotSupported
	for _, fn := range []func() error{
		func() error { _, err := p.ProbeQuality(context.Background(), "m"); return err },
		func() error { _, err := p.ProbeEmbedding(context.Background(), "m"); return err },
		func() error { _, err := p.ProbeStreaming(context.Background(), "m"); return err },
	} {
		if err := fn(); !errors.As(err, &ns) {
			t.Errorf("want ErrNotSupported, got %T: %v", err, err)
		}
	}
}

func TestParseMistralError(t *testing.T) {
	body := mustReadMistralFixture(t, "chat_completions_401.json")
	got := parseMistralError(body)
	if !strings.Contains(got, "Unauthorized") {
		t.Errorf("parseMistralError: got %q, want 'Unauthorized' substring", got)
	}
}
