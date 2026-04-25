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

func mustReadGeminiFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("gemini", "testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return b
}

func TestGemini_Identity(t *testing.T) {
	p := NewGeminiProvider("fake-key", "node-1")
	if got := p.ID(); got != "google_gemini" {
		t.Errorf("ID: got %q, want google_gemini", got)
	}
	if got := p.Models(); len(got) == 0 || got[0] != "gemini-2.5-flash" {
		t.Errorf("Models: got %v", got)
	}
}

func TestGemini_ProbeLightInference(t *testing.T) {
	cases := []struct {
		name         string
		httpStatus   int
		fixture      string
		wantSuccess  bool
		wantErrClass probes.ErrorClass
	}{
		{"success", http.StatusOK, "generate_content_200.json", true, probes.ErrorClassNone},
		{"auth", http.StatusUnauthorized, "generate_content_401.json", false, probes.ErrorClassAuth},
		{"rate_limit", http.StatusTooManyRequests, "generate_content_429.json", false, probes.ErrorClassRateLimit},
		{"server_5xx", http.StatusInternalServerError, "generate_content_500.json", false, probes.ErrorClassServer5xx},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := mustReadGeminiFixture(t, tc.fixture)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Gemini path includes model name and action.
				if !strings.Contains(r.URL.Path, "generateContent") {
					t.Errorf("path %q missing generateContent", r.URL.Path)
				}
				w.WriteHeader(tc.httpStatus)
				_, _ = w.Write(body)
			}))
			t.Cleanup(srv.Close)

			p := NewGeminiProvider("fake-key", "node-1", WithGeminiBaseURL(srv.URL))
			r, err := p.ProbeLightInference(context.Background(), "gemini-2.5-flash")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if r.ProviderID != "google_gemini" {
				t.Errorf("ProviderID: got %q, want google_gemini", r.ProviderID)
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

func TestGemini_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	p := NewGeminiProvider("fake-key", "node-1",
		WithGeminiBaseURL(srv.URL),
		WithGeminiHTTPClient(&http.Client{Timeout: 30 * time.Millisecond}),
	)
	r, err := p.ProbeLightInference(context.Background(), "gemini-2.5-flash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.ErrorClass != probes.ErrorClassTimeout {
		t.Errorf("ErrorClass: got %q, want timeout", r.ErrorClass)
	}
}

func TestGemini_UnsupportedProbes(t *testing.T) {
	p := NewGeminiProvider("fake-key", "node-1")
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

func TestParseGeminiError(t *testing.T) {
	body := mustReadGeminiFixture(t, "generate_content_429.json")
	got := parseGeminiError(body)
	if !strings.Contains(got, "exhausted") {
		t.Errorf("parseGeminiError: got %q, want 'exhausted' substring", got)
	}
}
