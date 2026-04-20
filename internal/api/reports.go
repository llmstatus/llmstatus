package api

import (
	"crypto/sha256"
	"fmt"
	"net"
	"net/http"
	"time"

	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

// postReport handles POST /v1/providers/{id}/report.
// It inserts a report for the provider keyed by the SHA-256 of the client IP.
// The SQL query enforces a 5-minute per-IP per-provider dedup; when the insert
// is a no-op (duplicate), we still return 204 so the client cannot probe
// whether another user reported from the same IP.
func (s *Server) postReport(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	// Verify provider exists; return 404 for unknown IDs.
	if _, err := s.store.GetProvider(r.Context(), id); err != nil {
		writeError(w, http.StatusNotFound, "provider not found")
		return
	}

	ipHash := hashIP(r)
	if err := s.store.InsertUserReport(r.Context(), pgstore.InsertUserReportParams{
		ProviderID: id,
		IpHash:     ipHash,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "could not record report")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// getReportHistogram handles GET /v1/providers/{id}/reports/histogram.
// Returns 24 hourly buckets (oldest first) with zero-filled gaps.
func (s *Server) getReportHistogram(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if _, err := s.store.GetProvider(r.Context(), id); err != nil {
		writeError(w, http.StatusNotFound, "provider not found")
		return
	}

	rows, err := s.store.UserReportHistogram(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not fetch histogram")
		return
	}

	type bucket struct {
		Hour  time.Time `json:"hour"`
		Count int64     `json:"count"`
	}
	out := make([]bucket, 0, len(rows))
	for _, row := range rows {
		// generate_series returns pgtype.Timestamptz at runtime; assert to time.Time.
		t, _ := row.Bucket.(time.Time)
		out = append(out, bucket{Hour: t.UTC(), Count: row.Count})
	}
	writeEnvelope(w, out)
}

// hashIP returns the hex-encoded SHA-256 of the client IP address (port stripped).
// We never store raw IPs.
func hashIP(r *http.Request) string {
	addr := r.RemoteAddr
	// Prefer X-Forwarded-For when behind a trusted proxy (nginx sets it).
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		addr = fwd
	}
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}
	sum := sha256.Sum256([]byte(host))
	return fmt.Sprintf("%x", sum)
}
