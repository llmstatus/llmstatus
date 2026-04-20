package api

import (
	"crypto/sha256"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

// postReport handles POST /v1/providers/{id}/report.
// It inserts a report for the provider keyed by the SHA-256 of the client IP.
// The SQL query enforces a 5-minute per-IP per-provider dedup; when the insert
// is a no-op (duplicate), we still return 204 so the client cannot probe
// whether another user reported from the same IP.
func (s *Server) postReport(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ipHash := hashIP(r)
	if err := s.store.InsertUserReport(r.Context(), pgstore.InsertUserReportParams{
		ProviderID: id,
		IpHash:     ipHash,
	}); err != nil {
		// FK violation means unknown provider_id — treat as 404.
		writeError(w, http.StatusNotFound, "provider not found")
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
		var t time.Time
		switch v := row.Bucket.(type) {
		case time.Time:
			t = v
		case pgtype.Timestamptz:
			t = v.Time
		}
		out = append(out, bucket{Hour: t.UTC(), Count: row.Count})
	}
	writeEnvelope(w, out)
}

// hashIP returns the hex-encoded SHA-256 of the client IP address (port stripped).
// We never store raw IPs.
//
// When behind nginx, X-Forwarded-For is "client, proxy1, ..., $remote_addr".
// We take only the rightmost entry — the one nginx itself appended — which cannot
// be spoofed by the client. Left-hand entries are client-controlled and untrusted.
func hashIP(r *http.Request) string {
	addr := r.RemoteAddr
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		parts := strings.Split(fwd, ",")
		addr = strings.TrimSpace(parts[len(parts)-1])
	}
	if host, _, err := net.SplitHostPort(addr); err == nil {
		addr = host
	}
	sum := sha256.Sum256([]byte(addr))
	return fmt.Sprintf("%x", sum)
}
