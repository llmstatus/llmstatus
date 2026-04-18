package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const requestIDHeader = "X-Request-ID"

type contextKey string

const ctxKeyRequestID contextKey = "request_id"

// requestIDMiddleware ensures every request has an X-Request-ID.
// Accepts an incoming header (from a trusted proxy or client) or generates
// a new UUID v4. The resolved value is propagated on the response and into
// the request context for downstream logging.
func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(requestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}
		w.Header().Set(requestIDHeader, id)
		ctx := context.WithValue(r.Context(), ctxKeyRequestID, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// corsMiddleware adds permissive CORS headers for the public read-only API.
// Handles preflight OPTIONS requests inline (204, no body).
//
// All origins are allowed: the API is unauthenticated, GET-only, and public.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Access-Control-Allow-Origin", "*")
		h.Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		h.Set("Access-Control-Allow-Headers", "Accept, Content-Type, X-Request-ID")
		h.Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// accessLogMiddleware logs one structured line per request.
// Requests to /healthz are skipped to avoid uptime-probe noise.
func accessLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		rw := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)

		id, _ := r.Context().Value(ctxKeyRequestID).(string)
		slog.Info("api: request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", id,
			"remote_ip", realIP(r),
		)
	})
}

// statusRecorder wraps ResponseWriter to capture the written status code.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// applyMiddleware wraps handler with the production middleware stack.
// Execution order (outer-to-inner): accessLog → cors → requestID → handler.
func applyMiddleware(handler http.Handler) http.Handler {
	return accessLogMiddleware(
		corsMiddleware(
			requestIDMiddleware(handler),
		),
	)
}
