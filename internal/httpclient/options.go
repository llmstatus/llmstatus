package httpclient

import (
	"net/http"
	"time"
)

// Version is advertised in the User-Agent header.
//
// Kept as a single source of truth so any build-time injection (for
// example, setting this via -ldflags at release time) only needs to
// touch one file.
const Version = "0.0.0-dev"

// Options configures a new HTTP client created by New.
//
// All fields are optional; zero values resolve to the documented defaults.
type Options struct {
	// Timeout bounds the total time of any single request (including
	// connection, TLS, reading headers, and reading the body).
	//
	// Default: 30s. Set explicitly to 0 to disable (not recommended).
	Timeout time.Duration

	// UserAgent overrides the default User-Agent header value.
	//
	// Default: "llmstatus.io/<Version>".
	UserAgent string

	// MaxRetries is the number of retry attempts for idempotent requests
	// (GET / HEAD) after the initial attempt.
	//
	// Default: 2.
	MaxRetries int

	// Transport is the base http.RoundTripper. The returned client wraps
	// this in middleware that injects User-Agent and X-Request-ID
	// headers.
	//
	// Default: http.DefaultTransport clone.
	Transport http.RoundTripper
}

func (o Options) withDefaults() Options {
	if o.Timeout == 0 {
		o.Timeout = 30 * time.Second
	}
	if o.UserAgent == "" {
		o.UserAgent = "llmstatus.io/" + Version
	}
	if o.MaxRetries == 0 {
		o.MaxRetries = 2
	}
	if o.Transport == nil {
		o.Transport = http.DefaultTransport.(*http.Transport).Clone()
	}
	return o
}
