package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/llmstatus/llmstatus/internal/api"
	"github.com/llmstatus/llmstatus/internal/store/influx"
	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

func TestListProviders_Empty(t *testing.T) {
	store := &fakeStore{}
	srv := api.New(store)

	rec := doGet(t, srv, "/v1/providers")

	if rec.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rec.Code)
	}
	var resp map[string]json.RawMessage
	mustDecode(t, rec.Body, &resp)
	// data must be [] not null
	if string(resp["data"]) != "[]" {
		t.Errorf("data: got %s, want []", resp["data"])
	}
}

func TestListProviders_Operational(t *testing.T) {
	store := &fakeStore{
		providers: []pgstore.Provider{fixtureProvider("openai", "OpenAI")},
	}
	liveReader := &fakeLiveStatsReader{
		stats: []influx.ProviderLiveStat{
			{ProviderID: "openai", Uptime24h: 0.99, P95Ms: 200},
		},
	}
	srv := api.New(store, api.WithLiveStatsReader(liveReader))

	rec := doGet(t, srv, "/v1/providers")

	var body struct {
		Data []struct {
			ID            string  `json:"id"`
			CurrentStatus string  `json:"current_status"`
			ActiveInc     *string `json:"active_incident_id"`
		} `json:"data"`
	}
	mustDecode(t, rec.Body, &body)

	if len(body.Data) != 1 {
		t.Fatalf("data length: got %d, want 1", len(body.Data))
	}
	if body.Data[0].ID != "openai" {
		t.Errorf("ID: got %q, want openai", body.Data[0].ID)
	}
	if body.Data[0].CurrentStatus != "operational" {
		t.Errorf("status: got %q, want operational", body.Data[0].CurrentStatus)
	}
	if body.Data[0].ActiveInc != nil {
		t.Error("expected no active incident")
	}
}

func TestListProviders_WithOngoingIncident(t *testing.T) {
	inc := fixtureIncident("openai", "2026-04-18-openai-down", "ongoing", "critical")
	store := &fakeStore{
		providers: []pgstore.Provider{fixtureProvider("openai", "OpenAI")},
		incidents: []pgstore.Incident{inc},
	}
	liveReader := &fakeLiveStatsReader{
		stats: []influx.ProviderLiveStat{
			{ProviderID: "openai", Uptime24h: 0.40, P95Ms: 1200},
		},
	}
	srv := api.New(store, api.WithLiveStatsReader(liveReader))

	rec := doGet(t, srv, "/v1/providers")

	var body struct {
		Data []struct {
			CurrentStatus    string  `json:"current_status"`
			ActiveIncidentID *string `json:"active_incident_id"`
		} `json:"data"`
	}
	mustDecode(t, rec.Body, &body)

	if body.Data[0].CurrentStatus != "down" {
		t.Errorf("status: got %q, want down", body.Data[0].CurrentStatus)
	}
	if body.Data[0].ActiveIncidentID == nil {
		t.Error("expected active_incident_id to be set")
	}
}

func TestListProviders_WithLiveStats(t *testing.T) {
	store := &fakeStore{
		providers: []pgstore.Provider{fixtureProvider("openai", "OpenAI")},
	}
	liveReader := &fakeLiveStatsReader{
		stats: []influx.ProviderLiveStat{
			{ProviderID: "openai", Uptime24h: 0.99, P95Ms: 450.5},
		},
	}
	srv := api.New(store, api.WithLiveStatsReader(liveReader))

	rec := doGet(t, srv, "/v1/providers")

	var body struct {
		Data []struct {
			ID        string   `json:"id"`
			Uptime24h *float64 `json:"uptime_24h"`
			P95Ms     *float64 `json:"p95_ms"`
		} `json:"data"`
	}
	mustDecode(t, rec.Body, &body)

	if len(body.Data) != 1 {
		t.Fatalf("data length: got %d, want 1", len(body.Data))
	}
	p := body.Data[0]
	if p.Uptime24h == nil {
		t.Fatal("uptime_24h: got nil, want non-nil")
	}
	if *p.Uptime24h < 0.989 || *p.Uptime24h > 0.991 {
		t.Errorf("uptime_24h: got %f, want ~0.99", *p.Uptime24h)
	}
	if p.P95Ms == nil {
		t.Fatal("p95_ms: got nil, want non-nil")
	}
	if *p.P95Ms != 450.5 {
		t.Errorf("p95_ms: got %f, want 450.5", *p.P95Ms)
	}
}

func TestListProviders_LiveStatsNil_OmitsFields(t *testing.T) {
	// When liveStats reader is configured but returns no data for a provider,
	// that provider is excluded from the response (no data = not yet probed).
	store := &fakeStore{
		providers: []pgstore.Provider{fixtureProvider("openai", "OpenAI")},
	}
	srv := api.New(store) // no WithLiveStatsReader → provider excluded

	rec := doGet(t, srv, "/v1/providers")

	var body struct {
		Data []json.RawMessage `json:"data"`
	}
	mustDecode(t, rec.Body, &body)
	// Provider has no live data, so it is filtered out entirely.
	if len(body.Data) != 0 {
		t.Fatalf("data length: got %d, want 0 (provider excluded when no live stats)", len(body.Data))
	}
}

func TestGetProvider_Found(t *testing.T) {
	store := &fakeStore{
		providers: []pgstore.Provider{fixtureProvider("anthropic", "Anthropic")},
		models:    []pgstore.Model{fixtureModel("anthropic", "claude-haiku-4-5-20251001")},
	}
	srv := api.New(store)

	rec := doGet(t, srv, "/v1/providers/anthropic")

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", rec.Code)
	}
	var body struct {
		Data struct {
			ID     string `json:"id"`
			Models []struct {
				ModelID string `json:"model_id"`
			} `json:"models"`
			ActiveIncidents []any `json:"active_incidents"`
		} `json:"data"`
	}
	mustDecode(t, rec.Body, &body)

	if body.Data.ID != "anthropic" {
		t.Errorf("ID: got %q, want anthropic", body.Data.ID)
	}
	if len(body.Data.Models) != 1 {
		t.Fatalf("models: got %d, want 1", len(body.Data.Models))
	}
	if body.Data.Models[0].ModelID != "claude-haiku-4-5-20251001" {
		t.Errorf("model_id: got %q", body.Data.Models[0].ModelID)
	}
	if body.Data.ActiveIncidents == nil {
		t.Error("active_incidents must not be null")
	}
}

func TestGetProvider_NotFound(t *testing.T) {
	store := &fakeStore{}
	srv := api.New(store)

	rec := doGet(t, srv, "/v1/providers/nonexistent")

	if rec.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want 404", rec.Code)
	}
}

func TestHealthz(t *testing.T) {
	srv := api.New(&fakeStore{})
	rec := doGet(t, srv, "/healthz")
	if rec.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rec.Code)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	srv := api.New(&fakeStore{})
	req := httptest.NewRequest(http.MethodPost, "/v1/providers", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	// Go 1.22 ServeMux returns 405 for wrong method on a known path.
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status: got %d, want 405", rec.Code)
	}
}

// ---- shared test helpers ----------------------------------------------------

func doGet(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func mustDecode(t *testing.T, body interface{ Read([]byte) (int, error) }, v any) {
	t.Helper()
	if err := json.NewDecoder(body).Decode(v); err != nil {
		t.Fatalf("decode: %v", err)
	}
}
