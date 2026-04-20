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

func mustReadDeepSeekFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("deepseek", "testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return b
}

func TestDeepSeek_Identity(t *testing.T) {
	p := NewDeepSeekProvider("sk-fake", "node-1")
	if got := p.ID(); got != "deepseek" {
		t.Errorf("ID: got %q, want deepseek", got)
	}
	if got := p.Models(); len(got) == 0 || got[0] != "deepseek-chat" {
		t.Errorf("Models: got %v", got)
	}
}

func TestDeepSeek_ProbeLightInference(t *testing.T) {
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
			body := mustReadDeepSeekFixture(t, tc.fixture)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/chat/completions" {
					t.Errorf("path: got %q, want /chat/completions", r.URL.Path)
				}
				w.WriteHeader(tc.httpStatus)
				_, _ = w.Write(body)
			}))
			t.Cleanup(srv.Close)

			p := NewDeepSeekProvider("sk-fake", "node-1", WithDeepSeekBaseURL(srv.URL))
			r, err := p.ProbeLightInference(context.Background(), "deepseek-chat")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if r.ProviderID != "deepseek" {
				t.Errorf("ProviderID: got %q, want deepseek", r.ProviderID)
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

func TestDeepSeek_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	p := NewDeepSeekProvider("sk-fake", "node-1",
		WithDeepSeekBaseURL(srv.URL),
		WithDeepSeekHTTPClient(&http.Client{Timeout: 30 * time.Millisecond}),
	)
	r, err := p.ProbeLightInference(context.Background(), "deepseek-chat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.ErrorClass != probes.ErrorClassTimeout {
		t.Errorf("ErrorClass: got %q, want timeout", r.ErrorClass)
	}
}

func TestDeepSeek_UnsupportedProbes(t *testing.T) {
	p := NewDeepSeekProvider("sk-fake", "node-1")
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
