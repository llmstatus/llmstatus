package adapters

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/llmstatus/llmstatus/internal/probes"
)

func TestOpenAI_Identity(t *testing.T) {
	p := NewOpenAIProvider("sk-fake", "node-1")
	if got := p.ID(); got != "openai" {
		t.Errorf("ID: got %q, want openai", got)
	}
	if got := p.Models(); len(got) == 0 || got[0] != "gpt-4o-mini" {
		t.Errorf("Models: got %v, want [gpt-4o-mini]", got)
	}
}

func TestOpenAI_ProbeLightInference(t *testing.T) {
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
		{"malformed_body", http.StatusOK, "chat_completions_malformed.html", false, probes.ErrorClassMalformedBody},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := mustReadFixture(t, tc.fixture)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if got := r.Header.Get("Authorization"); got != "Bearer sk-fake" {
					t.Errorf("Authorization header: got %q, want Bearer sk-fake", got)
				}
				if got := r.Header.Get("Content-Type"); got != "application/json" {
					t.Errorf("Content-Type: got %q, want application/json", got)
				}
				if r.URL.Path != "/chat/completions" {
					t.Errorf("path: got %q, want /chat/completions", r.URL.Path)
				}
				w.WriteHeader(tc.httpStatus)
				_, _ = w.Write(body)
			}))
			t.Cleanup(srv.Close)

			p := NewOpenAIProvider("sk-fake", "node-1", WithOpenAIBaseURL(srv.URL))
			r, err := p.ProbeLightInference(context.Background(), "gpt-4o-mini")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if r.ProviderID != "openai" {
				t.Errorf("ProviderID: got %q, want openai", r.ProviderID)
			}
			if r.ProbeType != "light_inference" {
				t.Errorf("ProbeType: got %q, want light_inference", r.ProbeType)
			}
			if r.RegionID != "node-1" {
				t.Errorf("RegionID: got %q, want node-1", r.RegionID)
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

func TestOpenAI_ProbeLightInference_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	client := &http.Client{Timeout: 30 * time.Millisecond}
	p := NewOpenAIProvider("sk-fake", "node-1",
		WithOpenAIBaseURL(srv.URL),
		WithOpenAIHTTPClient(client),
	)
	r, err := p.ProbeLightInference(context.Background(), "gpt-4o-mini")
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

func TestOpenAI_ProbeLightInference_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	p := NewOpenAIProvider("sk-fake", "node-1", WithOpenAIBaseURL(srv.URL))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	r, err := p.ProbeLightInference(ctx, "gpt-4o-mini")
	if err != nil {
		t.Fatalf("ProbeLightInference returned error: %v", err)
	}
	if r.ErrorClass != probes.ErrorClassTimeout {
		t.Errorf("ErrorClass: got %q, want timeout", r.ErrorClass)
	}
}

func TestOpenAI_ProbeQuality(t *testing.T) {
	cases := []struct {
		name         string
		httpStatus   int
		fixture      string
		wantSuccess  bool
		wantErrClass probes.ErrorClass
	}{
		{"success", http.StatusOK, "chat_completions_quality_200.json", true, probes.ErrorClassNone},
		{"mismatch", http.StatusOK, "chat_completions_quality_mismatch_200.json", false, probes.ErrorClassQualityMismatch},
		{"auth", http.StatusUnauthorized, "chat_completions_401.json", false, probes.ErrorClassAuth},
		{"rate_limit", http.StatusTooManyRequests, "chat_completions_429.json", false, probes.ErrorClassRateLimit},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := mustReadFixture(t, tc.fixture)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/chat/completions" {
					t.Errorf("path: got %q, want /chat/completions", r.URL.Path)
				}
				if r.Header.Get("User-Agent") != probeUserAgent {
					t.Errorf("User-Agent: got %q, want %q", r.Header.Get("User-Agent"), probeUserAgent)
				}
				w.WriteHeader(tc.httpStatus)
				_, _ = w.Write(body)
			}))
			t.Cleanup(srv.Close)

			p := NewOpenAIProvider("sk-fake", "node-1", WithOpenAIBaseURL(srv.URL))
			r, err := p.ProbeQuality(context.Background(), "gpt-4o-mini")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if r.ProbeType != "quality" {
				t.Errorf("ProbeType: got %q, want quality", r.ProbeType)
			}
			if r.Success != tc.wantSuccess {
				t.Errorf("Success: got %v, want %v", r.Success, tc.wantSuccess)
			}
			if r.ErrorClass != tc.wantErrClass {
				t.Errorf("ErrorClass: got %q, want %q", r.ErrorClass, tc.wantErrClass)
			}
		})
	}
}

func TestOpenAI_ProbeEmbedding(t *testing.T) {
	cases := []struct {
		name         string
		httpStatus   int
		fixture      string
		wantSuccess  bool
		wantErrClass probes.ErrorClass
	}{
		{"success", http.StatusOK, "embeddings_200.json", true, probes.ErrorClassNone},
		{"auth", http.StatusUnauthorized, "chat_completions_401.json", false, probes.ErrorClassAuth},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := mustReadFixture(t, tc.fixture)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/embeddings" {
					t.Errorf("path: got %q, want /embeddings", r.URL.Path)
				}
				w.WriteHeader(tc.httpStatus)
				_, _ = w.Write(body)
			}))
			t.Cleanup(srv.Close)

			p := NewOpenAIProvider("sk-fake", "node-1", WithOpenAIBaseURL(srv.URL))
			r, err := p.ProbeEmbedding(context.Background(), "gpt-4o-mini")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if r.ProbeType != "embedding" {
				t.Errorf("ProbeType: got %q, want embedding", r.ProbeType)
			}
			if r.Model != openaiEmbeddingModel {
				t.Errorf("Model: got %q, want %q", r.Model, openaiEmbeddingModel)
			}
			if r.Success != tc.wantSuccess {
				t.Errorf("Success: got %v, want %v", r.Success, tc.wantSuccess)
			}
			if r.ErrorClass != tc.wantErrClass {
				t.Errorf("ErrorClass: got %q, want %q", r.ErrorClass, tc.wantErrClass)
			}
			if tc.wantSuccess && r.TokensIn == 0 {
				t.Error("expected non-zero TokensIn on success")
			}
		})
	}
}

func TestOpenAI_ProbeStreaming(t *testing.T) {
	fixture := mustReadFixture(t, "chat_completions_stream_200.txt")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(fixture)
	}))
	t.Cleanup(srv.Close)

	p := NewOpenAIProvider("sk-fake", "node-1", WithOpenAIBaseURL(srv.URL))
	r, err := p.ProbeStreaming(context.Background(), "gpt-4o-mini")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.ProbeType != "streaming" {
		t.Errorf("ProbeType: got %q, want streaming", r.ProbeType)
	}
	if !r.Success {
		t.Errorf("expected Success=true, ErrorClass=%q", r.ErrorClass)
	}
	if r.DurationMs < 0 {
		t.Errorf("DurationMs negative: %d", r.DurationMs)
	}
}

func TestOpenAI_ProbeStreaming_Empty(t *testing.T) {
	// A stream that sends [DONE] without any content token is a failure.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	t.Cleanup(srv.Close)

	p := NewOpenAIProvider("sk-fake", "node-1", WithOpenAIBaseURL(srv.URL))
	r, err := p.ProbeStreaming(context.Background(), "gpt-4o-mini")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Success {
		t.Error("expected Success=false for empty stream")
	}
	if r.ErrorClass != probes.ErrorClassMalformedBody {
		t.Errorf("ErrorClass: got %q, want malformed_body", r.ErrorClass)
	}
}

func TestClassifyOpenAIStatus(t *testing.T) {
	cases := []struct {
		status int
		want   probes.ErrorClass
	}{
		{http.StatusUnauthorized, probes.ErrorClassAuth},
		{http.StatusForbidden, probes.ErrorClassAuth},
		{http.StatusTooManyRequests, probes.ErrorClassRateLimit},
		{http.StatusInternalServerError, probes.ErrorClassServer5xx},
		{http.StatusBadGateway, probes.ErrorClassServer5xx},
		{http.StatusBadRequest, probes.ErrorClassClient4xx},
		{http.StatusConflict, probes.ErrorClassClient4xx},
	}
	for _, c := range cases {
		if got := classifyOpenAIStatus(c.status); got != c.want {
			t.Errorf("classifyOpenAIStatus(%d): got %q, want %q", c.status, got, c.want)
		}
	}
}

func TestParseOpenAIError(t *testing.T) {
	body := mustReadFixture(t, "chat_completions_401.json")
	got := parseOpenAIError(body)
	if !strings.Contains(got, "Incorrect API key") {
		t.Errorf("parseOpenAIError: got %q, want substring 'Incorrect API key'", got)
	}

	// Body that is not valid JSON should fall back to raw body (truncated).
	bad := mustReadFixture(t, "chat_completions_malformed.html")
	got = parseOpenAIError(bad)
	if !strings.HasPrefix(got, "<!doctype") {
		t.Errorf("parseOpenAIError (malformed): got %q, want html prefix", got)
	}
	if len(got) > openaiErrorDetailMax {
		t.Errorf("parseOpenAIError: length %d exceeds max %d", len(got), openaiErrorDetailMax)
	}
}

func mustReadFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("openai", "testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return b
}
