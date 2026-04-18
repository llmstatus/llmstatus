package probes_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/llmstatus/llmstatus/internal/probes"
)

func TestHTTPSink_Receive_Success(t *testing.T) {
	var received bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: got %q, want POST", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type: got %q, want application/json", ct)
		}
		received = true
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	sink := probes.NewHTTPSink(srv.URL)
	err := sink.Receive(context.Background(), probes.ProbeResult{
		ProviderID: "openai",
		Model:      "gpt-4o-mini",
		ProbeType:  "light_inference",
		RegionID:   "us-west-2",
		StartedAt:  time.Now().UTC(),
		Success:    true,
		DurationMs: 312,
	})
	if err != nil {
		t.Fatalf("Receive: %v", err)
	}
	if !received {
		t.Error("expected server to receive the request")
	}
}

func TestHTTPSink_Receive_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"write failed"}`))
	}))
	t.Cleanup(srv.Close)

	sink := probes.NewHTTPSink(srv.URL)
	err := sink.Receive(context.Background(), probes.ProbeResult{
		ProviderID: "anthropic",
		Model:      "claude-haiku-4-5-20251001",
		ProbeType:  "light_inference",
		RegionID:   "eu-west-1",
		StartedAt:  time.Now().UTC(),
	})
	if err == nil {
		t.Error("expected error on 500 response, got nil")
	}
}
