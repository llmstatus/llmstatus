package api_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/llmstatus/llmstatus/internal/api"
)

func TestRateLimiter_Allow_UnderLimit(t *testing.T) {
	rl := api.NewRateLimiter(5, time.Minute)
	for i := range 5 {
		allowed, remaining, _ := rl.Allow("1.2.3.4")
		if !allowed {
			t.Fatalf("call %d: expected allowed", i+1)
		}
		if remaining != 5-1-i {
			t.Fatalf("call %d: remaining=%d want %d", i+1, remaining, 5-1-i)
		}
	}
}

func TestRateLimiter_Allow_OverLimit(t *testing.T) {
	rl := api.NewRateLimiter(3, time.Minute)
	for range 3 {
		rl.Allow("1.2.3.4") //nolint:errcheck
	}
	allowed, remaining, _ := rl.Allow("1.2.3.4")
	if allowed {
		t.Fatal("4th call should be denied")
	}
	if remaining != 0 {
		t.Fatalf("remaining=%d want 0", remaining)
	}
}

func TestRateLimiter_Allow_IsolatedIPs(t *testing.T) {
	rl := api.NewRateLimiter(1, time.Minute)
	allowed1, _, _ := rl.Allow("1.1.1.1")
	allowed2, _, _ := rl.Allow("2.2.2.2")
	if !allowed1 || !allowed2 {
		t.Fatal("each IP should get its own quota")
	}
}

func TestRateLimiter_Allow_WindowReset(t *testing.T) {
	rl := api.NewRateLimiter(1, 50*time.Millisecond)
	rl.Allow("1.2.3.4")
	allowed, _, _ := rl.Allow("1.2.3.4")
	if allowed {
		t.Fatal("second call in same window should be denied")
	}
	time.Sleep(60 * time.Millisecond)
	allowed, _, _ = rl.Allow("1.2.3.4")
	if !allowed {
		t.Fatal("first call after window reset should be allowed")
	}
}

func TestRateLimiter_Middleware_Headers(t *testing.T) {
	rl := api.NewRateLimiter(10, time.Minute)
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := rl.Middleware(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "5.5.5.5:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d want 200", rr.Code)
	}
	if v := rr.Header().Get("X-RateLimit-Limit"); v != "10" {
		t.Fatalf("X-RateLimit-Limit=%q want 10", v)
	}
	remaining, err := strconv.Atoi(rr.Header().Get("X-RateLimit-Remaining"))
	if err != nil || remaining != 9 {
		t.Fatalf("X-RateLimit-Remaining=%q want 9", rr.Header().Get("X-RateLimit-Remaining"))
	}
	if rr.Header().Get("X-RateLimit-Reset") == "" {
		t.Fatal("X-RateLimit-Reset header missing")
	}
}

func TestRateLimiter_Middleware_429(t *testing.T) {
	rl := api.NewRateLimiter(2, time.Minute)
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := rl.Middleware(next)

	for i := range 3 {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "9.9.9.9:999"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if i < 2 && rr.Code != http.StatusOK {
			t.Fatalf("call %d: status=%d want 200", i+1, rr.Code)
		}
		if i == 2 {
			if rr.Code != http.StatusTooManyRequests {
				t.Fatalf("3rd call: status=%d want 429", rr.Code)
			}
			if rr.Header().Get("Retry-After") == "" {
				t.Fatal("Retry-After header missing on 429")
			}
		}
	}
}

func TestRateLimiter_Middleware_XForwardedFor(t *testing.T) {
	rl := api.NewRateLimiter(1, time.Minute)
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := rl.Middleware(next)

	make2Requests := func(xff string) (int, int) {
		codes := [2]int{}
		for i := range 2 {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("X-Forwarded-For", xff)
			req.RemoteAddr = "10.0.0.1:0"
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			codes[i] = rr.Code
		}
		return codes[0], codes[1]
	}

	first, second := make2Requests("203.0.113.1")
	if first != 200 || second != 429 {
		t.Fatalf("XFF rate limit: got %d,%d want 200,429", first, second)
	}

	// Different XFF IP gets its own quota.
	first2, _ := make2Requests("203.0.113.2")
	if first2 != 200 {
		t.Fatalf("separate IP should be allowed: got %d", first2)
	}
}

func TestWithRateLimiter_Wired(t *testing.T) {
	rl := api.NewRateLimiter(1, time.Minute)
	srv := api.New(&fakeStore{}, api.WithRateLimiter(rl))

	// First request allowed.
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.RemoteAddr = "7.7.7.7:0"
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("first request: status=%d want 200", rr.Code)
	}

	// Second request from same IP denied.
	req2 := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req2.RemoteAddr = "7.7.7.7:0"
	rr2 := httptest.NewRecorder()
	srv.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request: status=%d want 429", rr2.Code)
	}
}

func TestServer_NoLimiter_NoRateLimit(t *testing.T) {
	srv := api.New(&fakeStore{}) // no WithRateLimiter
	for i := range 5 {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		req.RemoteAddr = fmt.Sprintf("8.8.8.8:%d", i)
		rr := httptest.NewRecorder()
		srv.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("request %d: status=%d want 200", i+1, rr.Code)
		}
	}
}
