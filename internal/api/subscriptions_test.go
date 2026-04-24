package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/llmstatus/llmstatus/internal/api"
	"github.com/llmstatus/llmstatus/internal/auth"
	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

// fakeAuthStore implements api.AuthStore for unit tests.
type fakeAuthStore struct {
	fakeStore
	subs      []pgstore.Subscription
	subRows   []pgstore.ListSubscriptionsByUserRow
	createErr error
	users     []pgstore.User
}

func (f *fakeAuthStore) UpsertUser(_ context.Context, e string) (pgstore.User, error) {
	return pgstore.User{ID: 1, Email: e}, nil
}
func (f *fakeAuthStore) GetUserByEmail(_ context.Context, email string) (pgstore.User, error) {
	for _, u := range f.users {
		if u.Email == email {
			return u, nil
		}
	}
	return pgstore.User{}, pgx.ErrNoRows
}
func (f *fakeAuthStore) GetUserByID(_ context.Context, id int64) (pgstore.User, error) {
	for _, u := range f.users {
		if u.ID == id {
			return u, nil
		}
	}
	return pgstore.User{}, pgx.ErrNoRows
}
func (f *fakeAuthStore) MarkUserVerified(_ context.Context, _ int64) error { return nil }
func (f *fakeAuthStore) CreateOTPToken(_ context.Context, _ pgstore.CreateOTPTokenParams) (pgstore.OtpToken, error) {
	return pgstore.OtpToken{}, nil
}
func (f *fakeAuthStore) ConsumeOTPToken(_ context.Context, _ pgstore.ConsumeOTPTokenParams) (pgstore.OtpToken, error) {
	return pgstore.OtpToken{}, nil
}
func (f *fakeAuthStore) UpsertOAuthAccount(_ context.Context, arg pgstore.UpsertOAuthAccountParams) (pgstore.OauthAccount, error) {
	return pgstore.OauthAccount{UserID: arg.UserID}, nil
}
func (f *fakeAuthStore) ListSubscriptionsByUser(_ context.Context, _ int64) ([]pgstore.ListSubscriptionsByUserRow, error) {
	return f.subRows, f.err
}
func (f *fakeAuthStore) GetSubscription(_ context.Context, id int64) (pgstore.Subscription, error) {
	for _, s := range f.subs {
		if s.ID == id {
			return s, nil
		}
	}
	return pgstore.Subscription{}, pgx.ErrNoRows
}
func (f *fakeAuthStore) CreateSubscription(_ context.Context, arg pgstore.CreateSubscriptionParams) (pgstore.Subscription, error) {
	if f.createErr != nil {
		return pgstore.Subscription{}, f.createErr
	}
	s := pgstore.Subscription{
		ID:          int64(len(f.subs) + 1),
		UserID:      arg.UserID,
		ProviderID:  arg.ProviderID,
		MinSeverity: arg.MinSeverity,
		EmailAlerts: arg.EmailAlerts,
		EmailDigest: arg.EmailDigest,
		WebhookUrl:  arg.WebhookUrl,
	}
	f.subs = append(f.subs, s)
	return s, nil
}
func (f *fakeAuthStore) UpdateSubscription(_ context.Context, arg pgstore.UpdateSubscriptionParams) (pgstore.Subscription, error) {
	for i, s := range f.subs {
		if s.ID == arg.ID && s.UserID == arg.UserID {
			f.subs[i].MinSeverity = arg.MinSeverity
			f.subs[i].EmailAlerts = arg.EmailAlerts
			f.subs[i].EmailDigest = arg.EmailDigest
			f.subs[i].WebhookUrl = arg.WebhookUrl
			return f.subs[i], nil
		}
	}
	return pgstore.Subscription{}, pgx.ErrNoRows
}
func (f *fakeAuthStore) DeleteSubscription(_ context.Context, arg pgstore.DeleteSubscriptionParams) error {
	for i, s := range f.subs {
		if s.ID == arg.ID && s.UserID == arg.UserID {
			f.subs = append(f.subs[:i], f.subs[i+1:]...)
			return nil
		}
	}
	return nil
}
func (f *fakeAuthStore) UpdateUserSettings(_ context.Context, _ pgstore.UpdateUserSettingsParams) error {
	return nil
}

const testJWTSecret = "test-secret-12345"

func newAuthServer(store *fakeAuthStore) *api.Server {
	return api.New(&fakeStore{}, api.WithAuth(&api.AuthConfig{
		Store:          store,
		JWTSecret:      testJWTSecret,
		InternalSecret: "internal",
	}))
}

//nolint:unparam
func bearerHeader(t *testing.T, userID int64, email string) string {
	t.Helper()
	tok, err := auth.SignJWT(userID, email, testJWTSecret)
	if err != nil {
		t.Fatalf("sign jwt: %v", err)
	}
	return "Bearer " + tok
}

func doRequest(t *testing.T, h http.Handler, method, path, bearer string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	if bearer != "" {
		req.Header.Set("Authorization", bearer)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestListSubscriptions_Unauthenticated(t *testing.T) {
	srv := newAuthServer(&fakeAuthStore{})
	rec := doRequest(t, srv, http.MethodGet, "/account/subscriptions", "", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("got %d, want 401", rec.Code)
	}
}

func TestListSubscriptions_Empty(t *testing.T) {
	srv := newAuthServer(&fakeAuthStore{})
	rec := doRequest(t, srv, http.MethodGet, "/account/subscriptions", bearerHeader(t, 1, "a@b.com"), nil)
	if rec.Code != http.StatusOK {
		t.Errorf("got %d, want 200", rec.Code)
	}
	var resp struct {
		Data []any `json:"data"`
	}
	mustDecode(t, rec.Body, &resp)
	if len(resp.Data) != 0 {
		t.Errorf("expected empty list, got %d items", len(resp.Data))
	}
}

func TestListSubscriptions_WithRows(t *testing.T) {
	store := &fakeAuthStore{
		subRows: []pgstore.ListSubscriptionsByUserRow{
			{ID: 1, UserID: 1, ProviderID: "openai", ProviderName: "OpenAI",
				MinSeverity: "major", EmailAlerts: true, EmailDigest: true,
				WebhookUrl: pgtype.Text{Valid: false}},
		},
	}
	srv := newAuthServer(store)
	rec := doRequest(t, srv, http.MethodGet, "/account/subscriptions", bearerHeader(t, 1, "a@b.com"), nil)
	if rec.Code != http.StatusOK {
		t.Errorf("got %d, want 200", rec.Code)
	}
	var resp struct {
		Data []struct {
			ProviderID   string `json:"provider_id"`
			ProviderName string `json:"provider_name"`
			RSSURL       string `json:"rss_url"`
		} `json:"data"`
	}
	mustDecode(t, rec.Body, &resp)
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 item, got %d", len(resp.Data))
	}
	if resp.Data[0].ProviderID != "openai" {
		t.Errorf("provider_id: got %q, want openai", resp.Data[0].ProviderID)
	}
	if resp.Data[0].RSSURL != "/v1/providers/openai/feed.xml" {
		t.Errorf("rss_url: got %q", resp.Data[0].RSSURL)
	}
}

func TestCreateSubscription_OK(t *testing.T) {
	store := &fakeAuthStore{}
	srv := newAuthServer(store)
	rec := doRequest(t, srv, http.MethodPost, "/account/subscriptions", bearerHeader(t, 1, "a@b.com"),
		map[string]any{"provider_id": "openai", "min_severity": "minor"})
	if rec.Code != http.StatusCreated {
		t.Errorf("got %d, want 201 — body: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data struct {
			ProviderID  string `json:"provider_id"`
			MinSeverity string `json:"min_severity"`
		} `json:"data"`
	}
	mustDecode(t, rec.Body, &resp)
	if resp.Data.ProviderID != "openai" {
		t.Errorf("provider_id: got %q", resp.Data.ProviderID)
	}
	if resp.Data.MinSeverity != "minor" {
		t.Errorf("min_severity: got %q", resp.Data.MinSeverity)
	}
}

func TestCreateSubscription_InvalidSeverity(t *testing.T) {
	srv := newAuthServer(&fakeAuthStore{})
	rec := doRequest(t, srv, http.MethodPost, "/account/subscriptions", bearerHeader(t, 1, "a@b.com"),
		map[string]any{"provider_id": "openai", "min_severity": "nuclear"})
	if rec.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rec.Code)
	}
}

func TestCreateSubscription_Conflict(t *testing.T) {
	store := &fakeAuthStore{createErr: errors.New("subscriptions_user_id_provider_id_key")}
	srv := newAuthServer(store)
	rec := doRequest(t, srv, http.MethodPost, "/account/subscriptions", bearerHeader(t, 1, "a@b.com"),
		map[string]any{"provider_id": "openai"})
	if rec.Code != http.StatusConflict {
		t.Errorf("got %d, want 409", rec.Code)
	}
}

func TestDeleteSubscription_OK(t *testing.T) {
	store := &fakeAuthStore{
		subs: []pgstore.Subscription{{ID: 5, UserID: 1, ProviderID: "openai"}},
	}
	srv := newAuthServer(store)
	rec := doRequest(t, srv, http.MethodDelete, "/account/subscriptions/5", bearerHeader(t, 1, "a@b.com"), nil)
	if rec.Code != http.StatusNoContent {
		t.Errorf("got %d, want 204", rec.Code)
	}
}

func TestDeleteSubscription_Unauthenticated(t *testing.T) {
	srv := newAuthServer(&fakeAuthStore{})
	rec := doRequest(t, srv, http.MethodDelete, "/account/subscriptions/1", "", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("got %d, want 401", rec.Code)
	}
}
