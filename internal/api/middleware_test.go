package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/llmstatus/llmstatus/internal/api"
	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

// ---- request ID ---------------------------------------------------------------

func TestRequestID_Generated(t *testing.T) {
	srv := api.New(&fakeStore{})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	id := rr.Header().Get("X-Request-ID")
	if id == "" {
		t.Fatal("X-Request-ID header missing from response")
	}
	// UUID v4: 36 chars, 4 hyphens
	if len(id) != 36 || strings.Count(id, "-") != 4 {
		t.Errorf("X-Request-ID %q doesn't look like a UUID", id)
	}
}

func TestRequestID_PropagatesIncoming(t *testing.T) {
	srv := api.New(&fakeStore{})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("X-Request-ID", "custom-id-12345")
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	if got := rr.Header().Get("X-Request-ID"); got != "custom-id-12345" {
		t.Errorf("X-Request-ID: got %q, want custom-id-12345", got)
	}
}

func TestRequestID_UniquePerRequest(t *testing.T) {
	srv := api.New(&fakeStore{})

	ids := make(map[string]bool)
	for range 5 {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rr := httptest.NewRecorder()
		srv.ServeHTTP(rr, req)
		id := rr.Header().Get("X-Request-ID")
		if ids[id] {
			t.Fatalf("duplicate request ID generated: %s", id)
		}
		ids[id] = true
	}
}

// ---- CORS ---------------------------------------------------------------------

func TestCORS_GetHasOriginHeader(t *testing.T) {
	srv := api.New(&fakeStore{})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("Access-Control-Allow-Origin: got %q, want *", got)
	}
}

func TestCORS_Preflight(t *testing.T) {
	srv := api.New(&fakeStore{})
	req := httptest.NewRequest(http.MethodOptions, "/v1/providers", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("preflight status=%d want 204", rr.Code)
	}
	if got := rr.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(got, "GET") {
		t.Errorf("Access-Control-Allow-Methods=%q missing GET", got)
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("Access-Control-Allow-Origin=%q want *", got)
	}
	if got := rr.Header().Get("Access-Control-Max-Age"); got == "" {
		t.Error("Access-Control-Max-Age header missing")
	}
	// X-Request-ID must be present even on preflight (requestID runs before cors).
	if got := rr.Header().Get("X-Request-ID"); got == "" {
		t.Error("X-Request-ID header missing on preflight response")
	}
	// Preflight should not forward to the actual handler.
	if rr.Body.Len() != 0 {
		t.Errorf("preflight response body should be empty, got %d bytes", rr.Body.Len())
	}
}

func TestCORS_AllEndpoints(t *testing.T) {
	store := &fakeStore{
		providers: []pgstore.Provider{fixtureProvider("openai", "OpenAI")},
	}
	srv := api.New(store)

	paths := []string{
		"/v1/providers",
		"/v1/incidents",
		"/feed.xml",
	}
	for _, path := range paths {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Host = "llmstatus.io"
		rr := httptest.NewRecorder()
		srv.ServeHTTP(rr, req)

		if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "*" {
			t.Errorf("%s: Access-Control-Allow-Origin=%q want *", path, got)
		}
	}
}

// ---- access logging -----------------------------------------------------------

func TestAccessLog_HealthzSkipped(t *testing.T) {
	// No way to assert "not logged" in unit tests without injecting a logger —
	// this test verifies /healthz still returns 200 with logging middleware active.
	srv := api.New(&fakeStore{})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("/healthz status=%d want 200", rr.Code)
	}
}

func TestAccessLog_StatusCaptured(t *testing.T) {
	// Verify that 404 from an unknown route is captured correctly (not swallowed).
	srv := api.New(&fakeStore{})
	req := httptest.NewRequest(http.MethodGet, "/v1/unknown-route", nil)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status=%d want 404", rr.Code)
	}
}

// ---- middleware ordering -------------------------------------------------------

func TestMiddlewareOrdering_RateLimitBeforeApplication(t *testing.T) {
	// Rate limiter fires at limit=1. Second request should be 429.
	// CORS and request-ID headers should still appear on the 429 response
	// because they are added by the outer middleware layers.
	rl := api.NewRateLimiter(1, 0) // window=0 means every call resets (use window=1min)
	rl2 := api.NewRateLimiter(1, 60_000_000_000) // 1 req per minute
	srv := api.New(&fakeStore{}, api.WithRateLimiter(rl2))

	req1 := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req1.RemoteAddr = "1.2.3.4:0"
	rr1 := httptest.NewRecorder()
	srv.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Fatalf("first request: %d", rr1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req2.RemoteAddr = "1.2.3.4:0"
	rr2 := httptest.NewRecorder()
	srv.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request: %d want 429", rr2.Code)
	}
	// CORS + request ID should still be set even on 429.
	if rr2.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("CORS header missing on 429 response")
	}
	if rr2.Header().Get("X-Request-ID") == "" {
		t.Error("X-Request-ID missing on 429 response")
	}
	_ = rl
}
