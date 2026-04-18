// Package api implements the public read API for llmstatus.io.
//
// All endpoints are GET, unauthenticated, and return JSON wrapped in a
// standard envelope:
//
//	{"data": ..., "meta": {"generated_at": "...", "cache_ttl_s": 30}}
package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

// compile-time check: pgstore.Queries satisfies Store.
var _ Store = (*pgstore.Queries)(nil)

// Server wires handlers to a ServeMux and owns the store reference.
type Server struct {
	store     Store
	history   HistoryReader   // optional; nil → GET /history returns 503
	liveStats LiveStatsReader // optional; nil → uptime24h/p95_ms omitted
	limiter   *RateLimiter    // optional; nil → no rate limiting
	mux       *http.ServeMux
	handler   http.Handler // mux optionally wrapped with limiter middleware
}

// New creates a Server and registers all routes. Pass functional options
// (e.g. WithHistoryReader, WithRateLimiter) to enable optional capabilities.
func New(store Store, opts ...func(*Server)) *Server {
	s := &Server{store: store, mux: http.NewServeMux()}
	for _, o := range opts {
		o(s)
	}
	s.registerRoutes()
	inner := http.Handler(s.mux)
	if s.limiter != nil {
		inner = s.limiter.Middleware(inner)
	}
	s.handler = applyMiddleware(inner)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /healthz", s.handleHealthz)
	s.mux.HandleFunc("GET /v1/providers", s.listProviders)
	s.mux.HandleFunc("GET /v1/providers/{id}", s.getProvider)
	s.mux.HandleFunc("GET /v1/providers/{id}/history", s.getProviderHistory)
	s.mux.HandleFunc("GET /v1/incidents", s.listIncidents)
	s.mux.HandleFunc("GET /v1/incidents/{id}", s.getIncident)
	s.mux.HandleFunc("GET /badge/{id}", s.getBadge)
	s.mux.HandleFunc("GET /feed.xml", s.getGlobalFeed)
	s.mux.HandleFunc("GET /v1/providers/{id}/feed.xml", s.getProviderFeed)
}

func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// ---- response helpers -------------------------------------------------------

const cacheTTL = 30

type meta struct {
	GeneratedAt time.Time `json:"generated_at"`
	CacheTTLS   int       `json:"cache_ttl_s"`
}

func writeMeta() meta { return meta{GeneratedAt: time.Now().UTC(), CacheTTLS: cacheTTL} }

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeEnvelope(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, map[string]any{
		"data": data,
		"meta": writeMeta(),
	})
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// coalesceSlice ensures nil slices become [] in JSON (never null).
func coalesceSlice[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

// ---- pgtype converters ------------------------------------------------------

func textVal(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

func timeVal(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	tm := t.Time.UTC()
	return &tm
}

func mustTime(t pgtype.Timestamptz) time.Time {
	return t.Time.UTC()
}
