package adapters

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/llmstatus/llmstatus/internal/probes"
)

func mustReadAnthropicFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile("anthropic/testdata/" + name)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return b
}

func TestAnthropic_Identity(t *testing.T) {
	p := NewAnthropicProvider("sk-ant-fake", "node-1")
	if got := p.ID(); got != "anthropic" {
		t.Errorf("ID: got %q, want anthropic", got)
	}
	models := p.Models()
	if len(models) == 0 || models[0] != "claude-haiku-4-5-20251001" {
		t.Errorf("Models: got %v, want [claude-haiku-4-5-20251001]", models)
	}
}

func TestAnthropic_ProbeLightInference(t *testing.T) {
	cases := []struct {
		name         string
		httpStatus   int
		fixture      string
		wantSuccess  bool
		wantErrClass probes.ErrorClass
	}{
		{"success", http.StatusOK, "messages_200.json", true, probes.ErrorClassNone},
		{"auth", http.StatusUnauthorized, "messages_401.json", false, probes.ErrorClassAuth},
		{"rate_limit", http.StatusTooManyRequests, "messages_429.json", false, probes.ErrorClassRateLimit},
		{"server_5xx", http.StatusInternalServerError, "messages_500.json", false, probes.ErrorClassServer5xx},
		{"overloaded_529", anthropicStatusOverloaded, "messages_529.json", false, probes.ErrorClassServer5xx},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := mustReadAnthropicFixture(t, tc.fixture)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if got := r.Header.Get("x-api-key"); got != "sk-ant-fake" {
					t.Errorf("x-api-key: got %q, want sk-ant-fake", got)
				}
				if got := r.Header.Get("anthropic-version"); got != anthropicVersion {
					t.Errorf("anthropic-version: got %q, want %q", got, anthropicVersion)
				}
				if got := r.Header.Get("Content-Type"); got != "application/json" {
					t.Errorf("Content-Type: got %q, want application/json", got)
				}
				if r.URL.Path != "/messages" {
					t.Errorf("path: got %q, want /messages", r.URL.Path)
				}
				w.WriteHeader(tc.httpStatus)
				_, _ = w.Write(body)
			}))
			t.Cleanup(srv.Close)

			p := NewAnthropicProvider("sk-ant-fake", "node-eu",
				WithAnthropicBaseURL(srv.URL),
			)
			r, err := p.ProbeLightInference(context.Background(), "claude-haiku-4-5-20251001")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if r.ProviderID != "anthropic" {
				t.Errorf("ProviderID: got %q, want anthropic", r.ProviderID)
			}
			if r.ProbeType != "light_inference" {
				t.Errorf("ProbeType: got %q, want light_inference", r.ProbeType)
			}
			if r.RegionID != "node-eu" {
				t.Errorf("RegionID: got %q, want node-eu", r.RegionID)
			}
			if r.Success != tc.wantSuccess {
				t.Errorf("Success: got %v, want %v", r.Success, tc.wantSuccess)
			}
			if r.ErrorClass != tc.wantErrClass {
				t.Errorf("ErrorClass: got %q, want %q", r.ErrorClass, tc.wantErrClass)
			}
			if r.HTTPStatus != tc.httpStatus {
				t.Errorf("HTTPStatus: got %d, want %d", r.HTTPStatus, tc.httpStatus)
			}
			if r.DurationMs < 0 {
				t.Errorf("DurationMs negative: %d", r.DurationMs)
			}
			if tc.wantSuccess && r.TokensIn == 0 && r.TokensOut == 0 {
				t.Error("expected non-zero token counts on success")
			}
		})
	}
}

func TestAnthropic_ProbeLightInference_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	client := &http.Client{Timeout: 30 * time.Millisecond}
	p := NewAnthropicProvider("sk-ant-fake", "node-1",
		WithAnthropicBaseURL(srv.URL),
		WithAnthropicHTTPClient(client),
	)
	r, err := p.ProbeLightInference(context.Background(), "claude-haiku-4-5-20251001")
	if err != nil {
		t.Fatalf("ProbeLightInference returned error: %v", err)
	}
	if r.Success {
		t.Error("expected Success=false on timeout")
	}
	if r.ErrorClass != probes.ErrorClassTimeout {
		t.Errorf("ErrorClass: got %q, want timeout", r.ErrorClass)
	}
}

func TestAnthropic_ProbeLightInference_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p := NewAnthropicProvider("sk-ant-fake", "node-1", WithAnthropicBaseURL(srv.URL))
	r, err := p.ProbeLightInference(ctx, "claude-haiku-4-5-20251001")
	if err != nil {
		t.Fatalf("unexpected hard error: %v", err)
	}
	if r.Success {
		t.Error("expected Success=false on cancelled context")
	}
}

func TestAnthropic_UnsupportedProbes(t *testing.T) {
	p := NewAnthropicProvider("sk-ant-fake", "node-1")
	ctx := context.Background()

	if _, err := p.ProbeQuality(ctx, "m"); err == nil {
		t.Error("ProbeQuality: expected ErrNotSupported")
	}
	if _, err := p.ProbeEmbedding(ctx, "m"); err == nil {
		t.Error("ProbeEmbedding: expected ErrNotSupported")
	}
	if _, err := p.ProbeStreaming(ctx, "m"); err == nil {
		t.Error("ProbeStreaming: expected ErrNotSupported")
	}
}

func TestClassifyAnthropicStatus(t *testing.T) {
	cases := []struct {
		status int
		want   probes.ErrorClass
	}{
		{200, probes.ErrorClassNone},
		{401, probes.ErrorClassAuth},
		{403, probes.ErrorClassAuth},
		{429, probes.ErrorClassRateLimit},
		{500, probes.ErrorClassServer5xx},
		{529, probes.ErrorClassServer5xx},
		{400, probes.ErrorClassClient4xx},
		{0, probes.ErrorClassUnknown},
	}
	for _, tc := range cases {
		if tc.status == 200 {
			continue // 200 is handled by classifyAnthropicResponse, not classifyAnthropicStatus
		}
		got := classifyAnthropicStatus(tc.status)
		if got != tc.want {
			t.Errorf("classifyAnthropicStatus(%d) = %q, want %q", tc.status, got, tc.want)
		}
	}
}
