package influx

import (
	"context"
	"net/http"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/llmstatus/llmstatus/internal/probes"
)

func TestLineWriter_WriteProbeResult_Success(t *testing.T) {
	var gotBody string
	var gotAuth string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	w := NewWriter(Config{Host: srv.URL, Token: "mytoken", Database: "llmstatus"}, nil)
	t.Cleanup(func() { _ = w.Close() })

	result := probes.ProbeResult{
		ProviderID: "openai",
		Model:      "gpt-4o-mini",
		ProbeType:  "light_inference",
		RegionID:   "us-west-2",
		StartedAt:  time.Unix(1745000000, 0).UTC(),
		Success:    true,
		DurationMs: 312,
		HTTPStatus: 200,
		TokensIn:   8,
		TokensOut:  1,
	}

	if err := w.WriteProbeResult(context.Background(), result); err != nil {
		t.Fatalf("WriteProbeResult: %v", err)
	}

	if gotAuth != "Token mytoken" {
		t.Errorf("Authorization: got %q, want %q", gotAuth, "Token mytoken")
	}
	if !strings.HasPrefix(gotBody, "probes,") {
		t.Errorf("line protocol: got %q, want prefix probes,", gotBody)
	}
	for _, want := range []string{"provider_id=openai", "model=gpt-4o-mini", "success=true", "duration_ms=312i", "tokens_in=8i"} {
		if !strings.Contains(gotBody, want) {
			t.Errorf("line protocol missing %q in %q", want, gotBody)
		}
	}
}

func TestLineWriter_WriteProbeResult_InfluxError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"code":"unauthorized"}`))
	}))
	t.Cleanup(srv.Close)

	w := NewWriter(Config{Host: srv.URL, Token: "bad", Database: "db"}, nil)
	err := w.WriteProbeResult(context.Background(), probes.ProbeResult{
		ProviderID: "openai", Model: "m", ProbeType: "light_inference",
		RegionID: "us-east-1", StartedAt: time.Now(),
	})
	if err == nil {
		t.Error("expected error on 401 from InfluxDB, got nil")
	}
}

func TestToLineProtocol_Failure(t *testing.T) {
	r := probes.ProbeResult{
		ProviderID: "anthropic",
		Model:      "claude-haiku-4-5-20251001",
		ProbeType:  "light_inference",
		RegionID:   "eu-west-1",
		StartedAt:  time.Unix(1745000000, 500).UTC(),
		Success:    false,
		DurationMs: 30050,
		HTTPStatus: 529,
		ErrorClass: probes.ErrorClassServer5xx,
	}
	line := toLineProtocol(r)

	for _, want := range []string{
		"probes,",
		"provider_id=anthropic",
		"error_class=server_5xx",
		"success=false",
		"http_status=529i",
	} {
		if !strings.Contains(line, want) {
			t.Errorf("missing %q in line: %q", want, line)
		}
	}
	// tokens_in/out should be absent when zero
	if strings.Contains(line, "tokens_in") || strings.Contains(line, "tokens_out") {
		t.Errorf("unexpected token fields in failure result: %q", line)
	}
}

func TestEscapeTag(t *testing.T) {
	cases := []struct{ in, want string }{
		{"openai", "openai"},
		{"aws bedrock", `aws\ bedrock`},
		{"key=value", `key\=value`},
		{"a,b", `a\,b`},
	}
	for _, tc := range cases {
		if got := escapeTag(tc.in); got != tc.want {
			t.Errorf("escapeTag(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
