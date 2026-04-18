package detector

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

// ---- fake ProbeReader -------------------------------------------------------

type fakeReader struct {
	stats5m   []ProbeStats
	stats10m  []ProbeStats
	latency   []LatencyStats
	regional  []RegionalStats
	err       error
}

func (f *fakeReader) ErrorRateByProvider(_ context.Context, window time.Duration) ([]ProbeStats, error) {
	if f.err != nil {
		return nil, f.err
	}
	if window <= 5*time.Minute {
		return f.stats5m, nil
	}
	return f.stats10m, nil
}

func (f *fakeReader) LatencyByProvider(_ context.Context, _ time.Duration) ([]LatencyStats, error) {
	return f.latency, f.err
}

func (f *fakeReader) RegionalErrorRateByProvider(_ context.Context, _ time.Duration) ([]RegionalStats, error) {
	return f.regional, f.err
}

// ---- fake IncidentStore -----------------------------------------------------

type fakeIncidentStore struct {
	incidents []pgstore.Incident
	created   []pgstore.CreateIncidentParams
	resolved  []pgstore.ResolveIncidentParams
}

func (f *fakeIncidentStore) GetOngoingByProviderAndRule(_ context.Context, arg pgstore.GetOngoingByProviderAndRuleParams) (pgstore.Incident, error) {
	for _, inc := range f.incidents {
		if inc.ProviderID == arg.ProviderID &&
			inc.DetectionRule.Valid &&
			inc.DetectionRule.String == arg.DetectionRule.String &&
			inc.Status == "ongoing" {
			return inc, nil
		}
	}
	return pgstore.Incident{}, pgx.ErrNoRows
}

func (f *fakeIncidentStore) CreateIncident(_ context.Context, arg pgstore.CreateIncidentParams) (pgstore.Incident, error) {
	f.created = append(f.created, arg)
	inc := pgstore.Incident{
		ID:              uuid.New(),
		Slug:            arg.Slug,
		ProviderID:      arg.ProviderID,
		Severity:        arg.Severity,
		Title:           arg.Title,
		Status:          arg.Status,
		DetectionMethod: arg.DetectionMethod,
		DetectionRule:   arg.DetectionRule,
	}
	f.incidents = append(f.incidents, inc)
	return inc, nil
}

func (f *fakeIncidentStore) ResolveIncident(_ context.Context, arg pgstore.ResolveIncidentParams) error {
	f.resolved = append(f.resolved, arg)
	for i, inc := range f.incidents {
		if inc.ID == arg.ID {
			f.incidents[i].Status = "resolved"
		}
	}
	return nil
}

func (f *fakeIncidentStore) ListIncidentsByStatus(_ context.Context, arg pgstore.ListIncidentsByStatusParams) ([]pgstore.Incident, error) {
	var out []pgstore.Incident
	for _, inc := range f.incidents {
		if inc.Status == arg.Status {
			out = append(out, inc)
		}
	}
	return out, nil
}

// ---- tests ------------------------------------------------------------------

func TestRunner_CreatesIncident_WhenDown(t *testing.T) {
	reader := &fakeReader{
		stats5m: []ProbeStats{{ProviderID: "openai", Total: 10, Errors: 6}},
	}
	store := &fakeIncidentStore{}
	r := New(reader, store, time.Hour)
	r.runOnce(context.Background())

	if len(store.created) != 1 {
		t.Fatalf("expected 1 incident created, got %d", len(store.created))
	}
	inc := store.created[0]
	if inc.ProviderID != "openai" {
		t.Errorf("ProviderID: got %q, want openai", inc.ProviderID)
	}
	if inc.Severity != "critical" {
		t.Errorf("Severity: got %q, want critical", inc.Severity)
	}
	if inc.DetectionRule.String != RuleProviderDown {
		t.Errorf("DetectionRule: got %q, want %q", inc.DetectionRule.String, RuleProviderDown)
	}
	if inc.Status != "ongoing" {
		t.Errorf("Status: got %q, want ongoing", inc.Status)
	}
}

func TestRunner_Deduplication(t *testing.T) {
	// Pre-seed an existing ongoing incident.
	existing := pgstore.Incident{
		ID:              uuid.New(),
		ProviderID:      "openai",
		Status:          "ongoing",
		DetectionMethod: "auto",
		DetectionRule:   pgtype.Text{String: RuleProviderDown, Valid: true},
	}
	reader := &fakeReader{
		stats5m: []ProbeStats{{ProviderID: "openai", Total: 10, Errors: 6}},
	}
	store := &fakeIncidentStore{incidents: []pgstore.Incident{existing}}
	r := New(reader, store, time.Hour)
	r.runOnce(context.Background())

	if len(store.created) != 0 {
		t.Errorf("expected no new incidents (dedup), got %d", len(store.created))
	}
}

func TestRunner_ResolvesStaleIncident(t *testing.T) {
	// Ongoing incident for a rule that is no longer firing.
	stale := pgstore.Incident{
		ID:              uuid.New(),
		ProviderID:      "openai",
		Status:          "ongoing",
		DetectionMethod: "auto",
		DetectionRule:   pgtype.Text{String: RuleProviderDown, Valid: true},
	}
	reader := &fakeReader{} // no detections — all clear
	store := &fakeIncidentStore{incidents: []pgstore.Incident{stale}}
	r := New(reader, store, time.Hour)
	r.runOnce(context.Background())

	if len(store.resolved) != 1 {
		t.Fatalf("expected 1 resolution, got %d", len(store.resolved))
	}
	if store.resolved[0].ID != stale.ID {
		t.Errorf("resolved wrong incident")
	}
}

func TestRunner_DoesNotResolveManualIncidents(t *testing.T) {
	manual := pgstore.Incident{
		ID:              uuid.New(),
		ProviderID:      "openai",
		Status:          "ongoing",
		DetectionMethod: "manual",
		DetectionRule:   pgtype.Text{String: RuleProviderDown, Valid: true},
	}
	reader := &fakeReader{}
	store := &fakeIncidentStore{incidents: []pgstore.Incident{manual}}
	r := New(reader, store, time.Hour)
	r.runOnce(context.Background())

	if len(store.resolved) != 0 {
		t.Error("manual incident should not be auto-resolved")
	}
}

func TestRunner_ElevatedErrors_CreatesIncident(t *testing.T) {
	reader := &fakeReader{
		stats10m: []ProbeStats{{ProviderID: "anthropic", Total: 20, Errors: 2}}, // 10%
	}
	store := &fakeIncidentStore{}
	r := New(reader, store, time.Hour)
	r.runOnce(context.Background())

	if len(store.created) != 1 {
		t.Fatalf("expected 1 incident, got %d", len(store.created))
	}
	if store.created[0].Severity != "major" {
		t.Errorf("severity: got %q, want major", store.created[0].Severity)
	}
}

func TestIncidentSlug(t *testing.T) {
	t1 := time.Date(2026, 4, 18, 12, 0, 0, 0, time.UTC)
	got := incidentSlug("openai", RuleProviderDown, t1)
	want := "2026-04-18-openai-provider-down"
	if got != want {
		t.Errorf("slug: got %q, want %q", got, want)
	}
}
