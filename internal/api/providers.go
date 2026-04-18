package api

import (
	"net/http"
	"time"

	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

// ---- response types ---------------------------------------------------------

type providerSummary struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Category         string  `json:"category"`
	Region           string  `json:"region"`
	CurrentStatus    string  `json:"current_status"`
	ActiveIncidentID *string `json:"active_incident_id,omitempty"`
}

type modelSummary struct {
	ModelID     string `json:"model_id"`
	DisplayName string `json:"display_name"`
	ModelType   string `json:"model_type"`
	Active      bool   `json:"active"`
}

type incidentRef struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	Severity  string    `json:"severity"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	StartedAt time.Time `json:"started_at"`
}

type providerDetail struct {
	providerSummary
	StatusPageURL    *string       `json:"status_page_url,omitempty"`
	DocumentationURL *string       `json:"documentation_url,omitempty"`
	Models           []modelSummary `json:"models"`
	ActiveIncidents  []incidentRef  `json:"active_incidents"`
}

// ---- handlers ---------------------------------------------------------------

func (s *Server) listProviders(w http.ResponseWriter, r *http.Request) {
	providers, err := s.store.ListActiveProviders(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list providers")
		return
	}

	ongoing, err := s.store.ListIncidentsByStatus(r.Context(), pgstore.ListIncidentsByStatusParams{
		Status: "ongoing",
		Limit:  200,
		Offset: 0,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load incidents")
		return
	}

	// Build a map from provider_id → highest-severity ongoing incident.
	incidentByProvider := make(map[string]pgstore.Incident)
	for _, inc := range ongoing {
		existing, seen := incidentByProvider[inc.ProviderID]
		if !seen || severityRank(inc.Severity) > severityRank(existing.Severity) {
			incidentByProvider[inc.ProviderID] = inc
		}
	}

	summaries := make([]providerSummary, 0, len(providers))
	for _, p := range providers {
		s := toProviderSummary(p, incidentByProvider)
		summaries = append(summaries, s)
	}

	writeEnvelope(w, summaries)
}

func (s *Server) getProvider(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p, err := s.store.GetProvider(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "provider not found")
		return
	}

	models, err := s.store.ListModelsByProvider(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load models")
		return
	}

	activeIncs, err := s.store.ListIncidentsByProvider(r.Context(), pgstore.ListIncidentsByProviderParams{
		ProviderID: id,
		Limit:      20,
		Offset:     0,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load incidents")
		return
	}

	// Build incident map for status derivation.
	incMap := map[string]pgstore.Incident{}
	if len(activeIncs) > 0 {
		for _, inc := range activeIncs {
			if inc.Status == "ongoing" {
				existing, seen := incMap[inc.ProviderID]
				if !seen || severityRank(inc.Severity) > severityRank(existing.Severity) {
					incMap[inc.ProviderID] = inc
				}
			}
		}
	}

	detail := providerDetail{
		providerSummary:  toProviderSummary(p, incMap),
		StatusPageURL:    textVal(p.StatusPageUrl),
		DocumentationURL: textVal(p.DocumentationUrl),
		Models:           toModelSummaries(models),
		ActiveIncidents:  toIncidentRefs(activeIncs),
	}

	writeEnvelope(w, detail)
}

// ---- helpers ----------------------------------------------------------------

func toProviderSummary(p pgstore.Provider, incMap map[string]pgstore.Incident) providerSummary {
	status, activeID := deriveStatus(p.ID, incMap)
	return providerSummary{
		ID:               p.ID,
		Name:             p.Name,
		Category:         p.Category,
		Region:           p.Region,
		CurrentStatus:    status,
		ActiveIncidentID: activeID,
	}
}

func deriveStatus(providerID string, incMap map[string]pgstore.Incident) (string, *string) {
	inc, ok := incMap[providerID]
	if !ok {
		return "operational", nil
	}
	id := inc.ID.String()
	switch inc.Severity {
	case "critical":
		return "down", &id
	case "major":
		return "degraded", &id
	default:
		return "degraded", &id
	}
}

func severityRank(s string) int {
	switch s {
	case "critical":
		return 3
	case "major":
		return 2
	case "minor":
		return 1
	default:
		return 0
	}
}

func toModelSummaries(models []pgstore.Model) []modelSummary {
	out := make([]modelSummary, 0, len(models))
	for _, m := range models {
		out = append(out, modelSummary{
			ModelID:     m.ModelID,
			DisplayName: m.DisplayName,
			ModelType:   m.ModelType,
			Active:      m.Active,
		})
	}
	return coalesceSlice(out)
}

func toIncidentRefs(incidents []pgstore.Incident) []incidentRef {
	// Filter to ongoing only for the "active_incidents" field.
	out := make([]incidentRef, 0)
	for _, inc := range incidents {
		if inc.Status != "ongoing" {
			continue
		}
		out = append(out, incidentRef{
			ID:        inc.ID.String(),
			Slug:      inc.Slug,
			Severity:  inc.Severity,
			Title:     inc.Title,
			Status:    inc.Status,
			StartedAt: mustTime(inc.StartedAt),
		})
	}
	return coalesceSlice(out)
}
