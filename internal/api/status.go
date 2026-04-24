package api

import (
	"net/http"

	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

type statusCounts struct {
	Total       int `json:"total"`
	Operational int `json:"operational"`
	Degraded    int `json:"degraded"`
	Down        int `json:"down"`
}

type statusResponse struct {
	Status string       `json:"status"`
	Counts statusCounts `json:"counts"`
}

// getStatus returns the aggregate system status across all active providers.
// Status is the worst observed: "down" > "degraded" > "operational".
func (s *Server) getStatus(w http.ResponseWriter, r *http.Request) {
	providers, err := s.store.ListActiveProviders(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list providers")
		return
	}

	ongoing, err := s.store.ListIncidentsByStatus(r.Context(), pgstore.ListIncidentsByStatusParams{
		Status: statusOngoing,
		Limit:  200,
		Offset: 0,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load incidents")
		return
	}

	incidentByProvider := make(map[string]pgstore.Incident)
	for _, inc := range ongoing {
		existing, seen := incidentByProvider[inc.ProviderID]
		if !seen || severityRank(inc.Severity) > severityRank(existing.Severity) {
			incidentByProvider[inc.ProviderID] = inc
		}
	}

	counts := statusCounts{Total: len(providers)}
	worst := "operational"
	for _, p := range providers {
		st, _ := deriveStatus(p.ID, incidentByProvider)
		switch st {
		case statusDown:
			counts.Down++
			worst = statusDown
		case statusDegraded:
			counts.Degraded++
			if worst != statusDown {
				worst = statusDegraded
			}
		default:
			counts.Operational++
		}
	}

	writeEnvelope(w, statusResponse{Status: worst, Counts: counts})
}
