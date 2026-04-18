package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/llmstatus/llmstatus/internal/api"
	"github.com/llmstatus/llmstatus/internal/store/influx"
)

type fakeHistoryReader struct {
	buckets []influx.HistoryBucket
	err     error
}

func (f *fakeHistoryReader) ProviderHistory(_ context.Context, _ string, _, _ time.Duration) ([]influx.HistoryBucket, error) {
	return f.buckets, f.err
}

func TestGetProviderHistory_Success(t *testing.T) {
	now := time.Now().UTC().Truncate(24 * time.Hour)
	fake := &fakeHistoryReader{
		buckets: []influx.HistoryBucket{
			{Timestamp: now.Add(-48 * time.Hour), Total: 1440, Errors: 14, Uptime: 0.9903},
			{Timestamp: now.Add(-24 * time.Hour), Total: 1440, Errors: 0, Uptime: 1.0},
		},
	}

	srv := api.New(&fakeStore{}, api.WithHistoryReader(fake))
	req := httptest.NewRequest(http.MethodGet, "/v1/providers/openai/history?window=7d", nil)
	req.SetPathValue("id", "openai")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	var env struct {
		Data []influx.HistoryBucket `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(env.Data) != 2 {
		t.Errorf("expected 2 buckets, got %d", len(env.Data))
	}
}

func TestGetProviderHistory_NoReader(t *testing.T) {
	srv := api.New(&fakeStore{}) // no WithHistoryReader
	req := httptest.NewRequest(http.MethodGet, "/v1/providers/openai/history", nil)
	req.SetPathValue("id", "openai")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status: got %d, want 503", w.Code)
	}
}

func TestGetProviderHistory_EmptyCoalesced(t *testing.T) {
	fake := &fakeHistoryReader{buckets: nil}
	srv := api.New(&fakeStore{}, api.WithHistoryReader(fake))
	req := httptest.NewRequest(http.MethodGet, "/v1/providers/openai/history", nil)
	req.SetPathValue("id", "openai")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	var env struct {
		Data []influx.HistoryBucket `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Data == nil {
		t.Error("data should be [] not null")
	}
}
