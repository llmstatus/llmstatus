package api

import (
	"context"
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

	// Fetch models for each provider (N ≈ 7; 30 s cache makes this acceptable).
	modelsByProvider := make(map[string][]pgstore.Model)
	for _, p := range providers {
		if models, err := s.store.ListModelsByProvider(r.Context(), p.ID); err == nil {
			modelsByProvider[p.ID] = models
		}
	}

	live := s.loadAllLiveStats(r.Context())
	incByProvider := ongoingIncidentMap(ongoing)

	summaries := make([]providerSummary, 0, len(providers))
	for _, p := range providers {
		st, hasData := live.byProvider[p.ID]
		if !hasData {
			continue // skip providers with no probe data yet
		}
		ps := toProviderSummary(p, incByProvider)
		u, p95 := st[0], st[1]
		ps.Uptime24h = &u
		ps.P95Ms = &p95
		ps.ModelStats = buildModelStats(p.ID, modelsByProvider[p.ID], live.byModel, live.sparklines)
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

	modelByModel, sparklines := s.loadModelLiveStats(r.Context())
	ps := toProviderSummary(p, ongoingIncidentMap(activeIncs))
	ps.ModelStats = buildModelStats(id, models, modelByModel, sparklines)

	detail := providerDetail{
		providerSummary:  ps,
		StatusPageURL:    textVal(p.StatusPageUrl),
		DocumentationURL: textVal(p.DocumentationUrl),
		Models:           toModelSummaries(models),
		ActiveIncidents:  toIncidentRefs(activeIncs),
		RegionStats:      coalesceSlice(s.loadRegionStats(r.Context(), id)),
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

// allLiveStats groups the three InfluxDB stat maps used by listProviders.
type allLiveStats struct {
	byProvider map[string][2]float64 // provider_id → [uptime24h, p95_ms]
	byModel    map[string][2]float64 // "provider_id:model" → [uptime24h, p95_ms]
	sparklines map[string][]float64  // "provider_id:model" → 60 avg_ms buckets
}

// loadAllLiveStats fetches provider-level, model-level, and sparkline stats.
// All queries are non-fatal; missing data results in zero values.
func (s *Server) loadAllLiveStats(ctx context.Context) allLiveStats {
	out := allLiveStats{
		byProvider: make(map[string][2]float64),
		byModel:    make(map[string][2]float64),
		sparklines: make(map[string][]float64),
	}
	if s.liveStats == nil {
		return out
	}
	if ps, err := s.liveStats.AllProviderLiveStats(ctx); err == nil {
		for _, st := range ps {
			out.byProvider[st.ProviderID] = [2]float64{st.Uptime24h, st.P95Ms}
		}
	}
	if ms, err := s.liveStats.AllModelLiveStats(ctx); err == nil {
		for _, m := range ms {
			out.byModel[m.ProviderID+":"+m.Model] = [2]float64{m.Uptime24h, m.P95Ms}
		}
	}
	if sl, err := s.liveStats.AllModelSparklines(ctx); err == nil {
		out.sparklines = sl
	}
	return out
}

// loadModelLiveStats fetches model-level stats and sparklines for getProvider.
func (s *Server) loadModelLiveStats(ctx context.Context) (map[string][2]float64, map[string][]float64) {
	byModel := make(map[string][2]float64)
	sparklines := make(map[string][]float64)
	if s.liveStats == nil {
		return byModel, sparklines
	}
	if ms, err := s.liveStats.AllModelLiveStats(ctx); err == nil {
		for _, m := range ms {
			byModel[m.ProviderID+":"+m.Model] = [2]float64{m.Uptime24h, m.P95Ms}
		}
	}
	if sl, err := s.liveStats.AllModelSparklines(ctx); err == nil {
		sparklines = sl
	}
	return byModel, sparklines
}

// loadRegionStats fetches per-region 24 h stats for a single provider.
func (s *Server) loadRegionStats(ctx context.Context, providerID string) []regionStat {
	if s.liveStats == nil {
		return nil
	}
	rs, err := s.liveStats.ProviderRegionStats(ctx, providerID)
	if err != nil {
		return nil
	}
	out := make([]regionStat, 0, len(rs))
	for _, reg := range rs {
		out = append(out, regionStat{
			RegionID:  reg.RegionID,
			Uptime24h: reg.Uptime24h,
			P95Ms:     reg.P95Ms,
		})
	}
	return out
}
