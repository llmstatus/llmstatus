package api

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RateLimiter enforces a fixed-window per-IP request limit.
type RateLimiter struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	hits   map[string]*rlEntry
}

type rlEntry struct {
	count int
	reset time.Time
}

// NewRateLimiter creates a RateLimiter allowing limit requests per window.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		limit:  limit,
		window: window,
		hits:   make(map[string]*rlEntry),
	}
}

// Allow checks whether ip may make another request.
// It returns (allowed, remaining, reset time).
func (rl *RateLimiter) Allow(ip string) (bool, int, time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	e, ok := rl.hits[ip]
	if !ok || now.After(e.reset) {
		reset := now.Truncate(rl.window).Add(rl.window)
		rl.hits[ip] = &rlEntry{count: 1, reset: reset}
		return true, rl.limit - 1, reset
	}

	e.count++
	remaining := rl.limit - e.count
	if remaining < 0 {
		remaining = 0
	}
	return e.count <= rl.limit, remaining, e.reset
}

// Middleware wraps next, enforcing rate limits and writing standard headers.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := realIP(r)
		allowed, remaining, reset := rl.Allow(ip)

		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.limit))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", reset.Unix()))

		if !allowed {
			retryAfter := int(time.Until(reset).Seconds()) + 1
			w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
			writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// WithRateLimiter wraps the server's handler with per-IP rate limiting.
func WithRateLimiter(rl *RateLimiter) func(*Server) {
	return func(s *Server) { s.limiter = rl }
}

// realIP extracts the client IP from X-Forwarded-For (first entry, set by nginx)
// or falls back to the TCP remote address.
func realIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first address; trim whitespace and strip port if present.
		first := strings.TrimSpace(strings.SplitN(xff, ",", 2)[0])
		if ip := net.ParseIP(first); ip != nil {
			return ip.String()
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
