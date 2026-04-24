package api

import (
	"net/http"
	"time"

	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

// Incident and severity values used across multiple handlers.
const (
	statusOngoing     = "ongoing"
	statusOperational = "operational"
	statusDown        = "down"
	statusDegraded    = "degraded"
	severityMajor     = "major"
)

// ---- response types ---------------------------------------------------------

type modelStat struct {
	ModelID     string    `json:"model_id"`
	DisplayName string    `json:"display_name"`
	Uptime24h   float64   `json:"uptime_24h"`
	P95Ms       float64   `json:"p95_ms"`
	Sparkline   []float64 `json:"sparkline"` // 60 avg_ms values; 0 = no data
}

type providerSummary struct {
	ID               string      `json:"id"`
	Name             string      `json:"name"`
	Category         string      `json:"category"`
	Region           string      `json:"region"`
	ProbeScope       string      `json:"probe_scope"`
	CurrentStatus    string      `json:"current_status"`
	ActiveIncidentID *string     `json:"active_incident_id,omitempty"`
	Uptime24h        *float64    `json:"uptime_24h,omitempty"`
	P95Ms            *float64    `json:"p95_ms,omitempty"`
	ModelStats       []modelStat `json:"model_stats"`
}

type regionStat struct {
	RegionID  string  `json:"region_id"`
	Uptime24h float64 `json:"uptime_24h"` // 0–1
	P95Ms     float64 `json:"p95_ms"`
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
	StatusPageURL    *string        `json:"status_page_url,omitempty"`
	DocumentationURL *string        `json:"documentation_url,omitempty"`
	Models           []modelSummary `json:"models"`
	ActiveIncidents  []incidentRef  `json:"active_incidents"`
	RegionStats      []regionStat   `json:"region_stats"` // 24 h stats per probe region
}

// ---- handlers ---------------------------------------------------------------

func (s *Server) listProviders(w http.ResponseWriter, r *http.Request) {
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

	// Fetch models for each provider (N ≈ 7; 30 s cache makes this acceptable).
	modelsByProvider := make(map[string][]pgstore.Model)
	for _, p := range providers {
		if models, err := s.store.ListModelsByProvider(r.Context(), p.ID); err == nil {
			modelsByProvider[p.ID] = models
		}
	}

	// Fetch all InfluxDB stats in parallel (three queries, all non-fatal).
	var (
		liveByProvider = make(map[string][2]float64) // provider_id → [uptime, p95]
		modelLiveStats = make(map[string][2]float64) // "pid:model" → [uptime, p95]
		sparklines     = make(map[string][]float64)  // "pid:model" → [60]float64
	)
	if s.liveStats != nil {
		ctx := r.Context()
		if stats, err := s.liveStats.AllProviderLiveStats(ctx); err == nil {
			for _, st := range stats {
				liveByProvider[st.ProviderID] = [2]float64{st.Uptime24h, st.P95Ms}
			}
		}
		if mstats, err := s.liveStats.AllModelLiveStats(ctx); err == nil {
			for _, ms := range mstats {
				modelLiveStats[ms.ProviderID+":"+ms.Model] = [2]float64{ms.Uptime24h, ms.P95Ms}
			}
		}
		if sl, err := s.liveStats.AllModelSparklines(ctx); err == nil {
			sparklines = sl
		}
	}

	summaries := make([]providerSummary, 0, len(providers))
	for _, p := range providers {
		ps := toProviderSummary(p, incidentByProvider)
		if live, ok := liveByProvider[p.ID]; ok {
			u, p95 := live[0], live[1]
			ps.Uptime24h = &u
			ps.P95Ms = &p95
		}
		ps.ModelStats = buildModelStats(p.ID, modelsByProvider[p.ID], modelLiveStats, sparklines)
		summaries = append(summaries, ps)
	}

	writeEnvelope(w, summaries)
}

func buildModelStats(
	providerID string,
	models []pgstore.Model,
	liveStats map[string][2]float64,
	sparklines map[string][]float64,
) []modelStat {
	out := make([]modelStat, 0, len(models))
	for _, m := range models {
		if !m.Active {
			continue
		}
		key := providerID + ":" + m.ModelID
		live := liveStats[key]
		sl := sparklines[key]
		if sl == nil {
			sl = make([]float64, 60)
		}
		out = append(out, modelStat{
			ModelID:     m.ModelID,
			DisplayName: m.DisplayName,
			Uptime24h:   live[0],
			P95Ms:       live[1],
			Sparkline:   sl,
		})
	}
	return coalesceSlice(out)
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
			if inc.Status == statusOngoing {
				existing, seen := incMap[inc.ProviderID]
				if !seen || severityRank(inc.Severity) > severityRank(existing.Severity) {
					incMap[inc.ProviderID] = inc
				}
			}
		}
	}

	// Fetch model live stats (non-fatal if InfluxDB is unavailable).
	var (
		modelLiveStats = make(map[string][2]float64)
		sparklines     = make(map[string][]float64)
	)
	if s.liveStats != nil {
		ctx := r.Context()
		if mstats, err := s.liveStats.AllModelLiveStats(ctx); err == nil {
			for _, ms := range mstats {
				modelLiveStats[ms.ProviderID+":"+ms.Model] = [2]float64{ms.Uptime24h, ms.P95Ms}
			}
		}
		if sl, err := s.liveStats.AllModelSparklines(ctx); err == nil {
			sparklines = sl
		}
	}

	var regionStats []regionStat
	if s.liveStats != nil {
		if rs, err := s.liveStats.ProviderRegionStats(r.Context(), id); err == nil {
			regionStats = make([]regionStat, 0, len(rs))
			for _, reg := range rs {
				regionStats = append(regionStats, regionStat{
					RegionID:  reg.RegionID,
					Uptime24h: reg.Uptime24h,
					P95Ms:     reg.P95Ms,
				})
			}
		}
	}

	ps := toProviderSummary(p, incMap)
	ps.ModelStats = buildModelStats(id, models, modelLiveStats, sparklines)

	detail := providerDetail{
		providerSummary:  ps,
		StatusPageURL:    textVal(p.StatusPageUrl),
		DocumentationURL: textVal(p.DocumentationUrl),
		Models:           toModelSummaries(models),
		ActiveIncidents:  toIncidentRefs(activeIncs),
		RegionStats:      coalesceSlice(regionStats),
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
		ProbeScope:       p.ProbeScope,
		CurrentStatus:    status,
		ActiveIncidentID: activeID,
	}
}

func deriveStatus(providerID string, incMap map[string]pgstore.Incident) (string, *string) {
	inc, ok := incMap[providerID]
	if !ok {
		return statusOperational, nil
	}
	id := inc.ID.String()
	switch inc.Severity {
	case "critical":
		return statusDown, &id
	case severityMajor:
		return statusDegraded, &id
	default:
		return statusDegraded, &id
	}
}

func severityRank(s string) int {
	switch s {
	case "critical":
		return 3
	case severityMajor:
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
		if inc.Status != statusOngoing {
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
