// Package api implements the public read API for llmstatus.io.
//
// All endpoints are GET, unauthenticated, and return JSON wrapped in a
// standard envelope:
//
//	{"data": ..., "meta": {"generated_at": "...", "cache_ttl_s": 30}}
package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/llmstatus/llmstatus/internal/keyenc"
	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

// compile-time check: pgstore.Queries satisfies Store.
var _ Store = (*pgstore.Queries)(nil)

// Pinger is satisfied by *pgxpool.Pool; used by /healthz.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Server wires handlers to a ServeMux and owns the store reference.
type Server struct {
	store     Store
	history   HistoryReader   // optional; nil → GET /history returns 503
	liveStats LiveStatsReader // optional; nil → uptime24h/p95_ms omitted
	pinger    Pinger          // optional; nil → /healthz always 200
	limiter   *RateLimiter    // optional; nil → no rate limiting
	auth      *AuthConfig     // optional; nil → auth routes return 501
	keyEnc    *keyenc.Encrypter // optional; nil → sponsor key endpoints return 503
	mux       *http.ServeMux
	handler   http.Handler // mux optionally wrapped with limiter middleware
}

// WithPinger enables a real DB liveness check in /healthz.
func WithPinger(p Pinger) func(*Server) {
	return func(s *Server) { s.pinger = p }
}

// WithAuth enables auth routes (OTP, OAuth, JWT).
func WithAuth(cfg *AuthConfig) func(*Server) {
	return func(s *Server) { s.auth = cfg }
}

// WithKeyEncrypter enables sponsor API key storage.
func WithKeyEncrypter(enc *keyenc.Encrypter) func(*Server) {
	return func(s *Server) { s.keyEnc = enc }
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
	s.mux.HandleFunc("GET /v1/status", s.getStatus)
	s.mux.HandleFunc("GET /v1/providers", s.listProviders)
	s.mux.HandleFunc("GET /v1/providers/{id}", s.getProvider)
	s.mux.HandleFunc("GET /v1/providers/{id}/history", s.getProviderHistory)
	s.mux.HandleFunc("GET /v1/incidents", s.listIncidents)
	s.mux.HandleFunc("GET /v1/incidents/{id}", s.getIncident)
	s.mux.HandleFunc("GET /badge/{id}", s.getBadge)
	s.mux.HandleFunc("GET /feed.xml", s.getGlobalFeed)
	s.mux.HandleFunc("GET /v1/providers/{id}/feed.xml", s.getProviderFeed)

	if s.auth != nil {
		s.mux.HandleFunc("POST /auth/otp/send", s.handleOTPSend)
		s.mux.HandleFunc("POST /auth/otp/verify", s.handleOTPVerify)
		s.mux.HandleFunc("POST /auth/oauth/upsert", s.handleOAuthUpsert)
		s.mux.HandleFunc("GET /auth/me", s.handleMe)
		s.mux.HandleFunc("GET /account/subscriptions", s.handleListSubscriptions)
		s.mux.HandleFunc("POST /account/subscriptions", s.handleCreateSubscription)
		s.mux.HandleFunc("PUT /account/subscriptions/{id}", s.handleUpdateSubscription)
		s.mux.HandleFunc("DELETE /account/subscriptions/{id}", s.handleDeleteSubscription)
		s.mux.HandleFunc("PUT /account/settings", s.handleUpdateSettings)

		// Sponsor self-service
		s.mux.HandleFunc("GET /v1/sponsors", s.listSponsors)
		s.mux.HandleFunc("POST /v1/sponsor/register", s.registerSponsor)
		s.mux.HandleFunc("GET /v1/sponsor/me", s.getSponsorMe)
		s.mux.HandleFunc("PATCH /v1/sponsor/me", s.updateSponsorMe)
		s.mux.HandleFunc("PUT /v1/sponsor/me/keys/{provider_id}", s.upsertSponsorKey)
		s.mux.HandleFunc("DELETE /v1/sponsor/me/keys/{provider_id}", s.deleteSponsorKey)
	}
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if s.pinger != nil {
		if err := s.pinger.Ping(r.Context()); err != nil {
			writeError(w, http.StatusServiceUnavailable, "database unavailable")
			return
		}
	}
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

func mustTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t.UTC(), Valid: true}
}
