package api

import (
	"context"
	"net/http"
	"time"

	"github.com/llmstatus/llmstatus/internal/store/influx"
)

// HistoryReader is the subset of influx.HistoryReader used by the API.
type HistoryReader interface {
	ProviderHistory(ctx context.Context, providerID string, window, bucketSize time.Duration) ([]influx.HistoryBucket, error)
}

// compile-time check: influx.influxHistoryReader satisfies HistoryReader via
// the exported NewHistoryReader constructor.
var _ HistoryReader = (influx.HistoryReader)(nil)

// WithHistoryReader wires an InfluxDB history reader into the Server.
func WithHistoryReader(r HistoryReader) func(*Server) {
	return func(s *Server) { s.history = r }
}

func (s *Server) getProviderHistory(w http.ResponseWriter, r *http.Request) {
	if s.history == nil {
		writeError(w, http.StatusServiceUnavailable, "history not available")
		return
	}

	id := r.PathValue("id")
	window, bucketSize := parseWindowParam(r.URL.Query().Get("window"))

	buckets, err := s.history.ProviderHistory(r.Context(), id, window, bucketSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to query history")
		return
	}

	writeEnvelope(w, coalesceSlice(buckets))
}

// parseWindowParam maps the ?window= query string to (total window, bucket size).
// Defaults to 30d with daily buckets.
func parseWindowParam(window string) (time.Duration, time.Duration) {
	switch window {
	case "24h":
		return 24 * time.Hour, time.Hour
	case "7d":
		return 7 * 24 * time.Hour, 24 * time.Hour
	default: // "30d" and anything else
		return 30 * 24 * time.Hour, 24 * time.Hour
	}
}
