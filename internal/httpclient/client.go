// Package httpclient provides the single HTTP client that every adapter
// uses to talk to an upstream AI provider.
//
// It centralises concerns that must be uniform across providers: a
// sensible default timeout, a stable User-Agent, a per-request ID that
// ties our logs to the provider's logs, and conservative retry behaviour
// for idempotent requests only.
package httpclient

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"
)

// New returns an *http.Client configured with the given Options.
//
// The returned client is safe for concurrent use by multiple goroutines.
// Its Transport wraps opts.Transport so it cannot be mutated externally
// after construction.
func New(opts Options) *http.Client {
	opts = opts.withDefaults()
	return &http.Client{
		Timeout: opts.Timeout,
		Transport: &decoratedTransport{
			base:       opts.Transport,
			userAgent:  opts.UserAgent,
			maxRetries: opts.MaxRetries,
		},
	}
}

// NewRequestID returns a 16-hex-char random identifier suitable for the
// X-Request-ID header. Exposed so tests can assert request-id is
// propagated through retry attempts.
func NewRequestID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		// rand.Read only returns an error if the OS RNG is unavailable,
		// which is a process-level disaster. Return a stable sentinel so
		// callers still see a non-empty value; this is observably wrong
		// and will show up in logs.
		return "00000000deadbeef"
	}
	return hex.EncodeToString(b[:])
}

type decoratedTransport struct {
	base       http.RoundTripper
	userAgent  string
	maxRetries int
}

func (t *decoratedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", t.userAgent)
	}
	if req.Header.Get("X-Request-ID") == "" {
		req.Header.Set("X-Request-ID", NewRequestID())
	}

	if !isIdempotent(req.Method) {
		return t.base.RoundTrip(req)
	}
	return t.roundTripWithRetry(req)
}

func (t *decoratedTransport) roundTripWithRetry(req *http.Request) (*http.Response, error) {
	var (
		resp *http.Response
		err  error
	)
	for attempt := 0; attempt <= t.maxRetries; attempt++ {
		resp, err = t.base.RoundTrip(req)
		if err == nil && resp.StatusCode < 500 {
			return resp, nil
		}
		if attempt == t.maxRetries {
			break // return the final resp/err unchanged
		}
		if resp != nil {
			_ = resp.Body.Close()
		}
		// Linear backoff: attempt 0 → 100ms, 1 → 200ms, 2 → 300ms.
		// Respect the request's context so client-wide timeouts are not
		// silently extended by retry sleeps.
		backoff := time.Duration(attempt+1) * 100 * time.Millisecond
		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		case <-time.After(backoff):
		}
	}
	return resp, err
}

func isIdempotent(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}
