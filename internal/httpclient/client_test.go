package httpclient_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/llmstatus/llmstatus/internal/httpclient"
)

func TestNew_DefaultHeadersAppliedOncePerRequest(t *testing.T) {
	var gotUA, gotRID string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		gotRID = r.Header.Get("X-Request-ID")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	client := httpclient.New(httpclient.Options{})
	resp, err := client.Get(srv.URL)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	_ = resp.Body.Close()

	if gotUA == "" || !containsPrefix(gotUA, "llmstatus.io/") {
		t.Errorf("User-Agent header missing or wrong: %q", gotUA)
	}
	if len(gotRID) < 8 {
		t.Errorf("X-Request-ID header too short or empty: %q", gotRID)
	}
}

func TestNew_RetriesOnlyIdempotentRequests(t *testing.T) {
	var getCalls, postCalls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			atomic.AddInt32(&getCalls, 1)
			w.WriteHeader(http.StatusInternalServerError)
		case http.MethodPost:
			atomic.AddInt32(&postCalls, 1)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	t.Cleanup(srv.Close)

	client := httpclient.New(httpclient.Options{MaxRetries: 2, Timeout: 5 * time.Second})

	resp, err := client.Get(srv.URL)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	_ = resp.Body.Close()
	if got := atomic.LoadInt32(&getCalls); got != 3 {
		t.Errorf("GET: want 3 attempts (1 initial + 2 retries), got %d", got)
	}

	resp, err = client.Post(srv.URL, "application/json", http.NoBody)
	if err != nil {
		t.Fatalf("Post: %v", err)
	}
	_ = resp.Body.Close()
	if got := atomic.LoadInt32(&postCalls); got != 1 {
		t.Errorf("POST: want 1 attempt (no retry on non-idempotent), got %d", got)
	}
}

func TestNew_HonoursCallerSuppliedHeaders(t *testing.T) {
	var gotUA, gotRID string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		gotRID = r.Header.Get("X-Request-ID")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	req, err := http.NewRequest(http.MethodGet, srv.URL, http.NoBody)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set("User-Agent", "test-agent/1.0")
	req.Header.Set("X-Request-ID", "deadbeef")

	resp, err := httpclient.New(httpclient.Options{}).Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	_ = resp.Body.Close()

	if gotUA != "test-agent/1.0" {
		t.Errorf("User-Agent overwritten: got %q, want test-agent/1.0", gotUA)
	}
	if gotRID != "deadbeef" {
		t.Errorf("X-Request-ID overwritten: got %q, want deadbeef", gotRID)
	}
}

func TestNew_TimeoutBoundsTotalDuration(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	client := httpclient.New(httpclient.Options{Timeout: 50 * time.Millisecond})
	start := time.Now()
	resp, err := client.Get(srv.URL)
	elapsed := time.Since(start)

	if err == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		t.Fatalf("expected timeout error, got nil (elapsed %v)", elapsed)
	}
	if elapsed > 300*time.Millisecond {
		t.Errorf("timeout not enforced tightly: %v", elapsed)
	}
}

func containsPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
