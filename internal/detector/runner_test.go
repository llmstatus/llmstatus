package detector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

// ---- fake ProbeReader -------------------------------------------------------

type fakeReader struct {
	stats5m  []ProbeStats
	stats10m []ProbeStats
	latency  []LatencyStats
	regional []RegionalStats
	err      error
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
	incidents  []pgstore.Incident
	created    []pgstore.CreateIncidentParams
	resolved   []pgstore.ResolveIncidentParams
	getErr     error // non-nil → GetOngoingByProviderAndRule returns this error
	createErr  error // non-nil → CreateIncident returns this error
	listErr    error // non-nil → ListIncidentsByStatus returns this error
	resolveErr error // non-nil → ResolveIncident returns this error
}

func (f *fakeIncidentStore) GetOngoingByProviderAndRule(_ context.Context, arg pgstore.GetOngoingByProviderAndRuleParams) (pgstore.Incident, error) {
	if f.getErr != nil {
		return pgstore.Incident{}, f.getErr
	}
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
	if f.createErr != nil {
		return pgstore.Incident{}, f.createErr
	}
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
	if f.resolveErr != nil {
		return f.resolveErr
	}
	f.resolved = append(f.resolved, arg)
	for i, inc := range f.incidents {
		if inc.ID == arg.ID {
			f.incidents[i].Status = "resolved"
		}
	}
	return nil
}

func (f *fakeIncidentStore) ListIncidentsByStatus(_ context.Context, arg pgstore.ListIncidentsByStatusParams) ([]pgstore.Incident, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
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

func TestRunner_Run_CancelReturnsCtxErr(t *testing.T) {
	reader := &fakeReader{}
	store := &fakeIncidentStore{}
	r := New(reader, store, 10*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- r.Run(ctx) }()

	// Let at least one tick fire, then cancel.
	time.Sleep(30 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Run returned %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run did not return after context cancel")
	}
}

func TestRunner_ReadError_5m_AbortsRunOnce(t *testing.T) {
	errRead := errors.New("influx timeout")
	reader := &fakeReader{err: errRead}
	store := &fakeIncidentStore{}
	r := New(reader, store, time.Hour)
	r.runOnce(context.Background())

	// runOnce should return early without creating any incidents.
	if len(store.created) != 0 {
		t.Errorf("expected 0 incidents on read error, got %d", len(store.created))
	}
}

func TestRunner_EnsureIncident_StoreGetError(t *testing.T) {
	// GetOngoingByProviderAndRule returns a non-ErrNoRows error → no incident created.
	reader := &fakeReader{
		stats5m: []ProbeStats{{ProviderID: "openai", Total: 10, Errors: 6}},
	}
	store := &fakeIncidentStore{getErr: errors.New("db connection lost")}
	r := New(reader, store, time.Hour)
	r.runOnce(context.Background())

	if len(store.created) != 0 {
		t.Errorf("expected 0 incidents when GetOngoing errors, got %d", len(store.created))
	}
}

func TestRunner_EnsureIncident_CreateError(t *testing.T) {
	// CreateIncident fails → no entry in store.created.
	reader := &fakeReader{
		stats5m: []ProbeStats{{ProviderID: "openai", Total: 10, Errors: 6}},
	}
	store := &fakeIncidentStore{createErr: errors.New("unique violation")}
	r := New(reader, store, time.Hour)
	r.runOnce(context.Background())

	if len(store.created) != 0 {
		t.Errorf("expected 0 recorded creates when CreateIncident errors, got %d", len(store.created))
	}
}

func TestRunner_ResolveStale_ListError(t *testing.T) {
	reader := &fakeReader{}
	store := &fakeIncidentStore{listErr: errors.New("db timeout")}
	r := New(reader, store, time.Hour)
	r.runOnce(context.Background()) // must not panic

	if len(store.resolved) != 0 {
		t.Errorf("expected 0 resolutions when list errors, got %d", len(store.resolved))
	}
}

func TestRunner_ResolveStale_ResolveError(t *testing.T) {
	stale := pgstore.Incident{
		ID:              uuid.New(),
		ProviderID:      "openai",
		Status:          "ongoing",
		DetectionMethod: "auto",
		DetectionRule:   pgtype.Text{String: RuleProviderDown, Valid: true},
	}
	reader := &fakeReader{}
	store := &fakeIncidentStore{
		incidents:  []pgstore.Incident{stale},
		resolveErr: errors.New("write failed"),
	}
	r := New(reader, store, time.Hour)
	r.runOnce(context.Background()) // must not panic

	// resolveErr causes continue — no entries in store.resolved.
	if len(store.resolved) != 0 {
		t.Errorf("expected 0 resolved on error, got %d", len(store.resolved))
	}
}

func TestIncidentTitle_AllRules(t *testing.T) {
	cases := []struct {
		rule string
		want string
	}{
		{RuleProviderDown, "openai is experiencing major disruption"},
		{RuleElevatedErrors, "openai elevated errors detected"},
		{RuleLatencyDegradation, "openai latency degradation detected"},
		{RuleRegionalOutage, "openai regional outage detected"},
		{"custom_rule", "openai custom_rule"},
	}
	for _, tc := range cases {
		got := incidentTitle("openai", tc.rule)
		if got != tc.want {
			t.Errorf("incidentTitle(%q): got %q, want %q", tc.rule, got, tc.want)
		}
	}
}
