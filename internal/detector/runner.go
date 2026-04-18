package detector

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

// IncidentStore is the subset of pgstore.Querier used by the detector.
type IncidentStore interface {
	GetOngoingByProviderAndRule(ctx context.Context, arg pgstore.GetOngoingByProviderAndRuleParams) (pgstore.Incident, error)
	CreateIncident(ctx context.Context, arg pgstore.CreateIncidentParams) (pgstore.Incident, error)
	ResolveIncident(ctx context.Context, arg pgstore.ResolveIncidentParams) error
	ListIncidentsByStatus(ctx context.Context, arg pgstore.ListIncidentsByStatusParams) ([]pgstore.Incident, error)
}

// compile-time check: pgstore.Queries satisfies IncidentStore.
var _ IncidentStore = (*pgstore.Queries)(nil)

// Runner periodically evaluates detection rules and manages incidents.
type Runner struct {
	reader   ProbeReader
	store    IncidentStore
	interval time.Duration
}

// New creates a Runner with the given reader, store, and tick interval.
func New(reader ProbeReader, store IncidentStore, interval time.Duration) *Runner {
	return &Runner{reader: reader, store: store, interval: interval}
}

// Run evaluates rules immediately, then repeats every interval.
// Blocks until ctx is cancelled; returns ctx.Err().
func (r *Runner) Run(ctx context.Context) error {
	r.runOnce(ctx)

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			r.runOnce(ctx)
		}
	}
}

func (r *Runner) runOnce(ctx context.Context) {
	stats5m, err := r.reader.ErrorRateByProvider(ctx, 5*time.Minute)
	if err != nil {
		slog.Error("detector: read 5m stats", "err", err)
		return
	}
	stats10m, err := r.reader.ErrorRateByProvider(ctx, 10*time.Minute)
	if err != nil {
		slog.Error("detector: read 10m stats", "err", err)
		return
	}

	detections := EvaluateRules(stats5m, stats10m)

	for _, d := range detections {
		r.ensureIncident(ctx, d)
	}

	r.resolveStale(ctx, detections)
}

// ensureIncident creates a new incident for the detection if none is already
// ongoing for the same provider+rule pair (deduplication).
func (r *Runner) ensureIncident(ctx context.Context, d Detection) {
	_, err := r.store.GetOngoingByProviderAndRule(ctx, pgstore.GetOngoingByProviderAndRuleParams{
		ProviderID:    d.ProviderID,
		DetectionRule: pgtype.Text{String: d.Rule, Valid: true},
	})
	if err == nil {
		return // already ongoing — do nothing
	}
	if !isNotFound(err) {
		slog.Error("detector: check ongoing incident", "provider", d.ProviderID, "rule", d.Rule, "err", err)
		return
	}

	now := time.Now().UTC()
	snapshot, _ := json.Marshal(map[string]any{
		"error_rate":   fmt.Sprintf("%.4f", d.ErrorRate),
		"total_probes": d.TotalProbes,
		"detected_at":  now.Format(time.RFC3339),
	})

	if _, err := r.store.CreateIncident(ctx, pgstore.CreateIncidentParams{
		Slug:            incidentSlug(d.ProviderID, d.Rule, now),
		ProviderID:      d.ProviderID,
		Severity:        d.Severity,
		Title:           incidentTitle(d.ProviderID, d.Rule),
		Status:          "ongoing",
		AffectedModels:  []string{},
		AffectedRegions: []string{},
		StartedAt:       pgtype.Timestamptz{Time: now, Valid: true},
		DetectionMethod: "auto",
		DetectionRule:   pgtype.Text{String: d.Rule, Valid: true},
		MetricsSnapshot: json.RawMessage(snapshot),
	}); err != nil {
		slog.Error("detector: create incident", "provider", d.ProviderID, "rule", d.Rule, "err", err)
		return
	}

	slog.Info("detector: incident created",
		"provider", d.ProviderID,
		"rule", d.Rule,
		"severity", d.Severity,
		"error_rate", fmt.Sprintf("%.1f%%", d.ErrorRate*100),
	)
}

// resolveStale resolves auto-detected ongoing incidents whose rule is no
// longer firing.
func (r *Runner) resolveStale(ctx context.Context, active []Detection) {
	ongoing, err := r.store.ListIncidentsByStatus(ctx, pgstore.ListIncidentsByStatusParams{
		Status: "ongoing",
		Limit:  200,
		Offset: 0,
	})
	if err != nil {
		slog.Error("detector: list ongoing incidents", "err", err)
		return
	}

	activeSet := make(map[string]struct{}, len(active))
	for _, d := range active {
		activeSet[d.ProviderID+"|"+d.Rule] = struct{}{}
	}

	now := time.Now().UTC()
	for _, inc := range ongoing {
		if inc.DetectionMethod != "auto" || !inc.DetectionRule.Valid {
			continue
		}
		key := inc.ProviderID + "|" + inc.DetectionRule.String
		if _, firing := activeSet[key]; firing {
			continue
		}
		if err := r.store.ResolveIncident(ctx, pgstore.ResolveIncidentParams{
			ID:         inc.ID,
			ResolvedAt: pgtype.Timestamptz{Time: now, Valid: true},
		}); err != nil {
			slog.Error("detector: resolve incident", "id", inc.ID, "err", err)
			continue
		}
		slog.Info("detector: incident resolved", "provider", inc.ProviderID, "rule", inc.DetectionRule.String)
	}
}

// ---- helpers ----------------------------------------------------------------

func incidentSlug(providerID, rule string, t time.Time) string {
	suffix := strings.ReplaceAll(rule, "_", "-")
	return t.Format("2006-01-02") + "-" + providerID + "-" + suffix
}

func incidentTitle(providerID, rule string) string {
	switch rule {
	case RuleProviderDown:
		return providerID + " is experiencing major disruption"
	case RuleElevatedErrors:
		return providerID + " elevated errors detected"
	default:
		return providerID + " " + rule
	}
}

func isNotFound(err error) bool {
	return err == pgx.ErrNoRows
}
