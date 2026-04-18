package influx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHistoryReader_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v3/query_sql" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"bucket":"2026-04-17T00:00:00Z","total":1440,"errors":14},
			{"bucket":"2026-04-18T00:00:00Z","total":720,"errors":0}
		]`))
	}))
	t.Cleanup(srv.Close)

	reader := NewHistoryReader(HistoryReaderConfig{
		Host: srv.URL, Token: "t", Database: "db",
	}, nil)

	buckets, err := reader.ProviderHistory(context.Background(), "openai", 30*24*time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatalf("ProviderHistory: %v", err)
	}
	if len(buckets) != 2 {
		t.Fatalf("expected 2 buckets, got %d", len(buckets))
	}

	b0 := buckets[0]
	if b0.Total != 1440 || b0.Errors != 14 {
		t.Errorf("bucket 0: got total=%d errors=%d", b0.Total, b0.Errors)
	}
	wantUptime := 1.0 - 14.0/1440.0
	if b0.Uptime < wantUptime-0.001 || b0.Uptime > wantUptime+0.001 {
		t.Errorf("bucket 0 uptime: got %f, want ~%f", b0.Uptime, wantUptime)
	}

	b1 := buckets[1]
	if b1.Uptime != 1.0 {
		t.Errorf("bucket 1 uptime: got %f, want 1.0 (no errors)", b1.Uptime)
	}
}

func TestHistoryReader_EmptyResult(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	t.Cleanup(srv.Close)

	reader := NewHistoryReader(HistoryReaderConfig{Host: srv.URL, Token: "t", Database: "db"}, nil)
	buckets, err := reader.ProviderHistory(context.Background(), "openai", 7*24*time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(buckets) != 0 {
		t.Errorf("expected 0 buckets, got %d", len(buckets))
	}
}

func TestHistoryReader_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"code":"unauthorized"}`))
	}))
	t.Cleanup(srv.Close)

	reader := NewHistoryReader(HistoryReaderConfig{Host: srv.URL, Token: "bad", Database: "db"}, nil)
	_, err := reader.ProviderHistory(context.Background(), "openai", 24*time.Hour, time.Hour)
	if err == nil {
		t.Error("expected error on 401, got nil")
	}
}

func TestHistoryReader_InvalidProviderID(t *testing.T) {
	reader := NewHistoryReader(HistoryReaderConfig{Host: "http://x", Token: "t", Database: "db"}, nil)
	_, err := reader.ProviderHistory(context.Background(), "'; DROP TABLE probes; --", time.Hour, time.Hour)
	if err == nil {
		t.Error("expected error for invalid provider_id")
	}
}

func TestHistoryReader_UptimeZeroTotal(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"bucket":"2026-04-18T00:00:00Z","total":0,"errors":0}]`))
	}))
	t.Cleanup(srv.Close)

	reader := NewHistoryReader(HistoryReaderConfig{Host: srv.URL, Token: "t", Database: "db"}, nil)
	buckets, err := reader.ProviderHistory(context.Background(), "openai", 24*time.Hour, time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(buckets) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(buckets))
	}
	if buckets[0].Uptime != 1.0 {
		t.Errorf("uptime with zero total: got %f, want 1.0", buckets[0].Uptime)
	}
}
