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

func TestInfluxReader_LatencyByProvider(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"provider_id":"openai","p95_ms":450.5,"total":20},
			{"provider_id":"anthropic","p95_ms":820.0,"total":15}
		]`))
	}))
	t.Cleanup(srv.Close)

	reader := NewInfluxReader(InfluxReaderConfig{
		Host: srv.URL, Token: "t", Database: "db",
	}, nil)

	stats, err := reader.LatencyByProvider(context.Background(), 5*time.Minute)
	if err != nil {
		t.Fatalf("LatencyByProvider: %v", err)
	}
	if len(stats) != 2 {
		t.Fatalf("expected 2 stats, got %d", len(stats))
	}
	byProvider := make(map[string]LatencyStats)
	for _, s := range stats {
		byProvider[s.ProviderID] = s
	}
	if got := byProvider["openai"].P95Ms; got != 450.5 {
		t.Errorf("openai P95Ms: got %f, want 450.5", got)
	}
	if got := byProvider["openai"].SampleCount; got != 20 {
		t.Errorf("openai SampleCount: got %d, want 20", got)
	}
}

func TestInfluxReader_RegionalErrorRateByProvider(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"provider_id":"openai","region_id":"us-east-1","total":10,"errors":6},
			{"provider_id":"openai","region_id":"eu-central","total":10,"errors":0}
		]`))
	}))
	t.Cleanup(srv.Close)

	reader := NewInfluxReader(InfluxReaderConfig{
		Host: srv.URL, Token: "t", Database: "db",
	}, nil)

	stats, err := reader.RegionalErrorRateByProvider(context.Background(), 5*time.Minute)
	if err != nil {
		t.Fatalf("RegionalErrorRateByProvider: %v", err)
	}
	if len(stats) != 2 {
		t.Fatalf("expected 2 stats, got %d", len(stats))
	}
	byRegion := make(map[string]RegionalStats)
	for _, s := range stats {
		byRegion[s.Region] = s
	}
	us := byRegion["us-east-1"]
	if us.Total != 10 || us.Errors != 6 {
		t.Errorf("us-east-1: total=%d errors=%d", us.Total, us.Errors)
	}
	if us.ErrorRate() < 0.59 || us.ErrorRate() > 0.61 {
		t.Errorf("us-east-1 error rate: got %f, want ~0.60", us.ErrorRate())
	}
}

func TestInfluxReader_RegionalErrorRateByProvider_SkipsEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// One row has empty region_id — should be skipped.
		_, _ = w.Write([]byte(`[
			{"provider_id":"openai","region_id":"","total":5,"errors":1},
			{"provider_id":"openai","region_id":"us-east-1","total":5,"errors":1}
		]`))
	}))
	t.Cleanup(srv.Close)

	reader := NewInfluxReader(InfluxReaderConfig{
		Host: srv.URL, Token: "t", Database: "db",
	}, nil)

	stats, err := reader.RegionalErrorRateByProvider(context.Background(), 5*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stats) != 1 {
		t.Errorf("expected 1 stat (empty region skipped), got %d", len(stats))
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
