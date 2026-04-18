package api

import (
	"errors"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

// ---- response types ---------------------------------------------------------

type incidentResponse struct {
	ID              string     `json:"id"`
	Slug            string     `json:"slug"`
	ProviderID      string     `json:"provider_id"`
	Severity        string     `json:"severity"`
	Title           string     `json:"title"`
	Description     *string    `json:"description,omitempty"`
	Status          string     `json:"status"`
	AffectedModels  []string   `json:"affected_models"`
	AffectedRegions []string   `json:"affected_regions"`
	StartedAt       time.Time  `json:"started_at"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty"`
	DetectionMethod string     `json:"detection_method"`
	HumanReviewed   bool       `json:"human_reviewed"`
}

// ---- handlers ---------------------------------------------------------------

func (s *Server) listIncidents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	status := q.Get("status")
	if status == "" || status == "all" {
		status = ""
	}

	limit := parseIntParam(q.Get("limit"), 20, 1, 100)
	offset := parseIntParam(q.Get("offset"), 0, 0, math.MaxInt32)

	var (
		incidents []pgstore.Incident
		err       error
	)

	if status != "" {
		incidents, err = s.store.ListIncidentsByStatus(r.Context(), pgstore.ListIncidentsByStatusParams{
			Status: status,
			Limit:  limit,
			Offset: offset,
		})
	} else {
		incidents, err = s.store.ListIncidents(r.Context(), pgstore.ListIncidentsParams{
			Limit:  limit,
			Offset: offset,
		})
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list incidents")
		return
	}

	writeEnvelope(w, toIncidentResponses(incidents))
}

func (s *Server) getIncident(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var (
		inc pgstore.Incident
		err error
	)

	// Try UUID first; fall back to slug lookup.
	if uid, uErr := uuid.Parse(id); uErr == nil {
		inc, err = s.store.GetIncidentByID(r.Context(), uid)
	} else {
		inc, err = s.store.GetIncidentBySlug(r.Context(), id)
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "incident not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch incident")
		return
	}

	writeEnvelope(w, toIncidentResponse(inc))
}

// ---- helpers ----------------------------------------------------------------

func toIncidentResponse(inc pgstore.Incident) incidentResponse {
	return incidentResponse{
		ID:              inc.ID.String(),
		Slug:            inc.Slug,
		ProviderID:      inc.ProviderID,
		Severity:        inc.Severity,
		Title:           inc.Title,
		Description:     textVal(inc.Description),
		Status:          inc.Status,
		AffectedModels:  coalesceSlice(inc.AffectedModels),
		AffectedRegions: coalesceSlice(inc.AffectedRegions),
		StartedAt:       mustTime(inc.StartedAt),
		ResolvedAt:      timeVal(inc.ResolvedAt),
		DetectionMethod: inc.DetectionMethod,
		HumanReviewed:   inc.HumanReviewed,
	}
}

func toIncidentResponses(incidents []pgstore.Incident) []incidentResponse {
	out := make([]incidentResponse, 0, len(incidents))
	for _, inc := range incidents {
		out = append(out, toIncidentResponse(inc))
	}
	return coalesceSlice(out)
}

func parseIntParam(v string, def, min, max int32) int32 {
	if v == "" {
		return def
	}
	n, err := strconv.ParseInt(v, 10, 32)
	if err != nil || n < int64(min) {
		return min
	}
	if n > int64(max) {
		return max
	}
	return int32(n)
}
