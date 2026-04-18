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
	srv := api.New(store)

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
	srv := api.New(store)

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
	store := &fakeStore{
		providers: []pgstore.Provider{fixtureProvider("openai", "OpenAI")},
	}
	srv := api.New(store) // no WithLiveStatsReader

	rec := doGet(t, srv, "/v1/providers")

	var body struct {
		Data []json.RawMessage `json:"data"`
	}
	mustDecode(t, rec.Body, &body)
	if len(body.Data) != 1 {
		t.Fatalf("data length: got %d, want 1", len(body.Data))
	}
	// Fields must be absent (omitempty), not null.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body.Data[0], &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := raw["uptime_24h"]; ok {
		t.Error("uptime_24h should be absent when liveStats is nil")
	}
	if _, ok := raw["p95_ms"]; ok {
		t.Error("p95_ms should be absent when liveStats is nil")
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
