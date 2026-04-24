package notifier

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/llmstatus/llmstatus/internal/email"
	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

// fakeStore implements Store for unit tests.
type fakeStore struct {
	mu        sync.Mutex
	incidents []pgstore.Incident
	subs      []pgstore.ListSubscriptionsForProviderRow
	logged    []pgstore.LogAlertParams
	sent      map[string]bool
}

//nolint:unparam
func alertKey(subID int64, incID uuid.UUID, ch, ev string) string {
	return incID.String() + ":" + ch + ":" + ev
}

func (f *fakeStore) ListIncidentsUpdatedSince(_ context.Context, _ pgtype.Timestamptz) ([]pgstore.Incident, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.incidents, nil
}
func (f *fakeStore) ListSubscriptionsForProvider(_ context.Context, _ string) ([]pgstore.ListSubscriptionsForProviderRow, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.subs, nil
}
func (f *fakeStore) IsAlertSent(_ context.Context, arg pgstore.IsAlertSentParams) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.sent[alertKey(arg.SubscriptionID, arg.IncidentID, arg.Channel, arg.Event)], nil
}
func (f *fakeStore) LogAlert(_ context.Context, arg pgstore.LogAlertParams) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.logged = append(f.logged, arg)
	if f.sent == nil {
		f.sent = make(map[string]bool)
	}
	f.sent[alertKey(arg.SubscriptionID, arg.IncidentID, arg.Channel, arg.Event)] = true
	return nil
}

func incidentFixture(providerID, status, severity string) pgstore.Incident {
	now := time.Now().UTC()
	return pgstore.Incident{
		ID:         uuid.New(),
		Slug:       providerID + "-test",
		ProviderID: providerID,
		Severity:   severity,
		Title:      "Test incident",
		Status:     status,
		StartedAt:  pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:  pgtype.Timestamptz{Time: now, Valid: true},
	}
}

//nolint:unparam
func subFixture(providerID, minSev, userEmail string, emailAlerts bool, webhookURL string) pgstore.ListSubscriptionsForProviderRow {
	row := pgstore.ListSubscriptionsForProviderRow{
		ID:           1,
		UserID:       1,
		ProviderID:   providerID,
		MinSeverity:  minSev,
		EmailAlerts:  emailAlerts,
		EmailDigest:  true,
		UserEmail:    userEmail,
		ProviderName: providerID,
	}
	if webhookURL != "" {
		row.WebhookUrl = pgtype.Text{String: webhookURL, Valid: true}
	}
	return row
}

// newTestNotifier creates a Notifier wired to a fake Resend server.
// Returns the notifier and a pointer to the list of received requests.
func newTestNotifier(t *testing.T, store *fakeStore) *Notifier {
	t.Helper()
	resend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"ok"}`))
	}))
	t.Cleanup(resend.Close)
	return New(Config{
		Store:   store,
		Email:   email.NewWithBaseURL(resend.URL, "key", "from@test.com"),
		SiteURL: "https://test.local",
	})
}

func TestPoll_SendsEmailOnNewIncident(t *testing.T) {
	inc := incidentFixture("openai", "ongoing", "major")
	store := &fakeStore{
		incidents: []pgstore.Incident{inc},
		subs:      []pgstore.ListSubscriptionsForProviderRow{subFixture("openai", "major", "u@test.com", true, "")},
	}
	newTestNotifier(t, store).poll(context.Background(), time.Now().Add(-time.Minute))

	store.mu.Lock()
	logged := store.logged
	store.mu.Unlock()

	if len(logged) != 1 {
		t.Fatalf("logged %d alerts, want 1", len(logged))
	}
	if logged[0].Event != eventCreated {
		t.Errorf("event = %q, want %q", logged[0].Event, eventCreated)
	}
	if logged[0].Channel != "email" {
		t.Errorf("channel = %q, want email", logged[0].Channel)
	}
}

func TestPoll_SeverityBelowMinSkipped(t *testing.T) {
	inc := incidentFixture("openai", "ongoing", "minor")
	store := &fakeStore{
		incidents: []pgstore.Incident{inc},
		subs:      []pgstore.ListSubscriptionsForProviderRow{subFixture("openai", "major", "u@test.com", true, "")},
	}
	newTestNotifier(t, store).poll(context.Background(), time.Now().Add(-time.Minute))

	store.mu.Lock()
	n := len(store.logged)
	store.mu.Unlock()
	if n != 0 {
		t.Errorf("logged %d, want 0 (minor below major threshold)", n)
	}
}

func TestPoll_NoDuplicateOnSecondPoll(t *testing.T) {
	inc := incidentFixture("openai", "ongoing", "major")
	store := &fakeStore{
		incidents: []pgstore.Incident{inc},
		subs:      []pgstore.ListSubscriptionsForProviderRow{subFixture("openai", "minor", "u@test.com", true, "")},
	}
	n := newTestNotifier(t, store)
	ctx := context.Background()
	since := time.Now().Add(-time.Minute)
	n.poll(ctx, since)
	n.poll(ctx, since)

	store.mu.Lock()
	total := len(store.logged)
	store.mu.Unlock()
	if total != 1 {
		t.Errorf("logged %d across 2 polls, want 1 (dedup)", total)
	}
}

func TestPoll_ResolvedEventDistinctFromCreated(t *testing.T) {
	inc := incidentFixture("openai", "resolved", "major")
	inc.ResolvedAt = pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	store := &fakeStore{
		incidents: []pgstore.Incident{inc},
		subs:      []pgstore.ListSubscriptionsForProviderRow{subFixture("openai", "minor", "u@test.com", true, "")},
	}
	newTestNotifier(t, store).poll(context.Background(), time.Now().Add(-time.Minute))

	store.mu.Lock()
	logged := store.logged
	store.mu.Unlock()
	if len(logged) != 1 {
		t.Fatalf("logged %d, want 1", len(logged))
	}
	if logged[0].Event != eventResolved {
		t.Errorf("event = %q, want %q", logged[0].Event, eventResolved)
	}
}

func TestPoll_WebhookDelivered(t *testing.T) {
	var hits int
	wh := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits++
		w.WriteHeader(http.StatusOK)
	}))
	defer wh.Close()

	inc := incidentFixture("anthropic", "ongoing", "critical")
	store := &fakeStore{
		incidents: []pgstore.Incident{inc},
		subs:      []pgstore.ListSubscriptionsForProviderRow{subFixture("anthropic", "critical", "u@test.com", false, wh.URL)},
	}
	newTestNotifier(t, store).poll(context.Background(), time.Now().Add(-time.Minute))

	if hits != 1 {
		t.Errorf("webhook hit %d times, want 1", hits)
	}
	store.mu.Lock()
	n := len(store.logged)
	store.mu.Unlock()
	if n != 1 {
		t.Errorf("logged %d webhook alerts, want 1", n)
	}
}

func TestWebhookRetry(t *testing.T) {
	var hits int
	wh := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer wh.Close()

	payload := webhookPayload{Event: "test", Title: "t"}
	if err := deliverWebhook(context.Background(), wh.URL, payload); err != nil {
		t.Fatalf("unexpected error after retries: %v", err)
	}
	if hits != 3 {
		t.Errorf("expected 3 attempts, got %d", hits)
	}
}

func TestIncidentEvent(t *testing.T) {
	cases := []struct{ status, want string }{
		{"ongoing", eventCreated},
		{"resolved", eventResolved},
		{"investigating", ""},
		{"", ""},
	}
	for _, tc := range cases {
		if got := incidentEvent(tc.status); got != tc.want {
			t.Errorf("incidentEvent(%q) = %q, want %q", tc.status, got, tc.want)
		}
	}
}
