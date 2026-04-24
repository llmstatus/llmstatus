package notifier

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/llmstatus/llmstatus/internal/email"
	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

// digestStore embeds fakeStore and adds DigestStore methods.
type digestStore struct {
	fakeStore
	mu          sync.Mutex
	users       []pgstore.User
	digSubs     map[int64][]pgstore.ListDigestSubscriptionsRow
	recentInc   map[string][]pgstore.Incident
	logged      []pgstore.LogDigestParams
	alreadySent map[string]bool
}

func (s *digestStore) sentKey(userID int64, date pgtype.Date) string {
	return fmt.Sprintf("%d:%s", userID, date.Time.Format("2006-01-02"))
}

func (s *digestStore) ListUsersForDigest(_ context.Context) ([]pgstore.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.users, nil
}
func (s *digestStore) ListDigestSubscriptions(_ context.Context, userID int64) ([]pgstore.ListDigestSubscriptionsRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.digSubs[userID], nil
}
func (s *digestStore) ListRecentIncidentsByProvider(_ context.Context, providerID string) ([]pgstore.Incident, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.recentInc[providerID], nil
}
func (s *digestStore) IsDigestSent(_ context.Context, arg pgstore.IsDigestSentParams) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.alreadySent[s.sentKey(arg.UserID, arg.SentDate)], nil
}
func (s *digestStore) LogDigest(_ context.Context, arg pgstore.LogDigestParams) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logged = append(s.logged, arg)
	if s.alreadySent == nil {
		s.alreadySent = make(map[string]bool)
	}
	s.alreadySent[s.sentKey(arg.UserID, arg.SentDate)] = true
	return nil
}

func newDigestNotifier(t *testing.T, store *digestStore) *Notifier {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"ok"}`))
	}))
	t.Cleanup(srv.Close)
	return New(Config{
		Store:   store,
		Email:   email.NewWithBaseURL(srv.URL, "key", "from@test.com"),
		SiteURL: "https://test.local",
	})
}

//nolint:unparam
func userFixture(id int64, digestHour int, tz string) pgstore.User {
	return pgstore.User{
		ID:         id,
		Email:      fmt.Sprintf("user%d@test.com", id),
		DigestHour: int32(digestHour),
		Timezone:   tz,
	}
}

func TestDigest_SendsWhenHourMatches(t *testing.T) {
	// User in UTC with digest_hour=8; call sendDigests at 08:00 UTC.
	now := time.Date(2026, 4, 20, 8, 0, 0, 0, time.UTC)
	u := userFixture(1, 8, "UTC")
	store := &digestStore{
		users: []pgstore.User{u},
		digSubs: map[int64][]pgstore.ListDigestSubscriptionsRow{
			1: {{ID: 1, UserID: 1, ProviderID: "openai", ProviderName: "OpenAI", EmailDigest: true}},
		},
		recentInc: map[string][]pgstore.Incident{},
	}
	n := newDigestNotifier(t, store)
	n.sendDigests(context.Background(), now)

	store.mu.Lock()
	n2 := len(store.logged)
	store.mu.Unlock()
	if n2 != 1 {
		t.Errorf("logged %d digests, want 1", n2)
	}
}

func TestDigest_SkipsWhenHourDoesNotMatch(t *testing.T) {
	now := time.Date(2026, 4, 20, 9, 0, 0, 0, time.UTC) // 09:00 UTC, not 08:00
	u := userFixture(1, 8, "UTC")
	store := &digestStore{users: []pgstore.User{u}}
	n := newDigestNotifier(t, store)
	n.sendDigests(context.Background(), now)

	store.mu.Lock()
	n2 := len(store.logged)
	store.mu.Unlock()
	if n2 != 0 {
		t.Errorf("logged %d digests, want 0 (wrong hour)", n2)
	}
}

func TestDigest_NoDuplicateSameDay(t *testing.T) {
	now := time.Date(2026, 4, 20, 8, 0, 0, 0, time.UTC)
	u := userFixture(1, 8, "UTC")
	store := &digestStore{
		users: []pgstore.User{u},
		digSubs: map[int64][]pgstore.ListDigestSubscriptionsRow{
			1: {{ID: 1, UserID: 1, ProviderID: "openai", ProviderName: "OpenAI", EmailDigest: true}},
		},
		recentInc: map[string][]pgstore.Incident{},
	}
	n := newDigestNotifier(t, store)
	n.sendDigests(context.Background(), now)
	n.sendDigests(context.Background(), now) // second call same hour

	store.mu.Lock()
	n2 := len(store.logged)
	store.mu.Unlock()
	if n2 != 1 {
		t.Errorf("logged %d digests across 2 calls, want 1 (dedup)", n2)
	}
}

func TestDigest_TimezoneOffset(t *testing.T) {
	// Asia/Shanghai is UTC+8. At 00:00 UTC it's 08:00 CST.
	now := time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC)
	u := userFixture(1, 8, "Asia/Shanghai")
	store := &digestStore{
		users: []pgstore.User{u},
		digSubs: map[int64][]pgstore.ListDigestSubscriptionsRow{
			1: {{ID: 1, UserID: 1, ProviderID: "openai", ProviderName: "OpenAI", EmailDigest: true}},
		},
		recentInc: map[string][]pgstore.Incident{},
	}
	n := newDigestNotifier(t, store)
	n.sendDigests(context.Background(), now)

	store.mu.Lock()
	n2 := len(store.logged)
	store.mu.Unlock()
	if n2 != 1 {
		t.Errorf("logged %d digests, want 1 (Asia/Shanghai +8 = 08:00 local)", n2)
	}
}
