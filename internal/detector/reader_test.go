package detector

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestInfluxReader_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v3/query_sql" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if auth := r.Header.Get("Authorization"); auth != "Token testtoken" {
			t.Errorf("Authorization: got %q", auth)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"provider_id":"openai","total":10,"errors":3},
			{"provider_id":"anthropic","total":5,"errors":0}
		]`))
	}))
	t.Cleanup(srv.Close)

	reader := NewInfluxReader(InfluxReaderConfig{
		Host:     srv.URL,
		Token:    "testtoken",
		Database: "llmstatus",
	}, nil)

	stats, err := reader.ErrorRateByProvider(context.Background(), 5*time.Minute)
	if err != nil {
		t.Fatalf("ErrorRateByProvider: %v", err)
	}
	if len(stats) != 2 {
		t.Fatalf("expected 2 stats, got %d", len(stats))
	}

	byProvider := make(map[string]ProbeStats)
	for _, s := range stats {
		byProvider[s.ProviderID] = s
	}

	openai := byProvider["openai"]
	if openai.Total != 10 || openai.Errors != 3 {
		t.Errorf("openai stats: got total=%d errors=%d", openai.Total, openai.Errors)
	}
	if openai.ErrorRate() < 0.29 || openai.ErrorRate() > 0.31 {
		t.Errorf("openai error rate: got %f, want ~0.30", openai.ErrorRate())
	}
}

func TestInfluxReader_InfluxError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"code":"unauthorized"}`))
	}))
	t.Cleanup(srv.Close)

	reader := NewInfluxReader(InfluxReaderConfig{
		Host: srv.URL, Token: "bad", Database: "db",
	}, nil)

	_, err := reader.ErrorRateByProvider(context.Background(), 5*time.Minute)
	if err == nil {
		t.Error("expected error on 401, got nil")
	}
}

func TestInfluxReader_EmptyResult(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	t.Cleanup(srv.Close)

	reader := NewInfluxReader(InfluxReaderConfig{
		Host: srv.URL, Token: "t", Database: "db",
	}, nil)

	stats, err := reader.ErrorRateByProvider(context.Background(), 5*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stats) != 0 {
		t.Errorf("expected 0 stats, got %d", len(stats))
	}
}
