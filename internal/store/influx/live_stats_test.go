package influx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAllProviderLiveStats_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v3/query_sql" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"provider_id":"openai","total":1440,"errors":144,"p95_ms":750.0},
			{"provider_id":"anthropic","total":720,"errors":0,"p95_ms":0}
		]`))
	}))
	t.Cleanup(srv.Close)

	reader := NewHistoryReader(HistoryReaderConfig{
		Host: srv.URL, Token: "t", Database: "db",
	}, nil)

	stats, err := reader.AllProviderLiveStats(context.Background())
	if err != nil {
		t.Fatalf("AllProviderLiveStats: %v", err)
	}
	if len(stats) != 2 {
		t.Fatalf("expected 2 stats, got %d", len(stats))
	}

	openai := stats[0]
	if openai.ProviderID != "openai" {
		t.Errorf("provider_id: got %q, want openai", openai.ProviderID)
	}
	wantUptime := 1.0 - 144.0/1440.0
	if openai.Uptime24h < wantUptime-0.001 || openai.Uptime24h > wantUptime+0.001 {
		t.Errorf("uptime: got %f, want ~%f", openai.Uptime24h, wantUptime)
	}
	if openai.P95Ms != 750.0 {
		t.Errorf("p95_ms: got %f, want 750.0", openai.P95Ms)
	}

	anthropic := stats[1]
	if anthropic.Uptime24h != 1.0 {
		t.Errorf("anthropic uptime: got %f, want 1.0", anthropic.Uptime24h)
	}
}

func TestAllProviderLiveStats_Empty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	t.Cleanup(srv.Close)

	reader := NewHistoryReader(HistoryReaderConfig{
		Host: srv.URL, Token: "t", Database: "db",
	}, nil)

	stats, err := reader.AllProviderLiveStats(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stats) != 0 {
		t.Errorf("expected 0 stats, got %d", len(stats))
	}
}

func TestAllProviderLiveStats_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	reader := NewHistoryReader(HistoryReaderConfig{
		Host: srv.URL, Token: "t", Database: "db",
	}, nil)

	_, err := reader.AllProviderLiveStats(context.Background())
	if err == nil {
		t.Fatal("expected error on 500, got nil")
	}
}
