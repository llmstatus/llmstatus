package api_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/llmstatus/llmstatus/internal/api"
	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

func TestListIncidents_Empty(t *testing.T) {
	srv := api.New(&fakeStore{})
	rec := doGet(t, srv, "/v1/incidents")

	if rec.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rec.Code)
	}
	var resp map[string]json.RawMessage
	mustDecode(t, rec.Body, &resp)
	if string(resp["data"]) != "[]" {
		t.Errorf("data: got %s, want []", resp["data"])
	}
}

func TestListIncidents_All(t *testing.T) {
	store := &fakeStore{
		incidents: []pgstore.Incident{
			fixtureIncident("openai", "slug-1", "ongoing", "major"),
			fixtureIncident("anthropic", "slug-2", "resolved", "minor"),
		},
	}
	srv := api.New(store)

	var body struct {
		Data []struct {
			Slug   string `json:"slug"`
			Status string `json:"status"`
		} `json:"data"`
	}
	mustDecode(t, doGet(t, srv, "/v1/incidents").Body, &body)

	if len(body.Data) != 2 {
		t.Fatalf("data length: got %d, want 2", len(body.Data))
	}
}

func TestListIncidents_FilterByStatus(t *testing.T) {
	store := &fakeStore{
		incidents: []pgstore.Incident{
			fixtureIncident("openai", "slug-1", "ongoing", "major"),
			fixtureIncident("anthropic", "slug-2", "resolved", "minor"),
		},
	}
	srv := api.New(store)

	var body struct {
		Data []struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	mustDecode(t, doGet(t, srv, "/v1/incidents?status=ongoing").Body, &body)

	if len(body.Data) != 1 {
		t.Fatalf("data length: got %d, want 1", len(body.Data))
	}
	if body.Data[0].Status != "ongoing" {
		t.Errorf("status: got %q, want ongoing", body.Data[0].Status)
	}
}

func TestGetIncident_BySlug(t *testing.T) {
	inc := fixtureIncident("openai", "2026-04-18-openai-down", "ongoing", "critical")
	store := &fakeStore{incidents: []pgstore.Incident{inc}}
	srv := api.New(store)

	var body struct {
		Data struct {
			Slug            string `json:"slug"`
			ProviderID      string `json:"provider_id"`
			AffectedModels  []any  `json:"affected_models"`
			AffectedRegions []any  `json:"affected_regions"`
		} `json:"data"`
	}
	mustDecode(t, doGet(t, srv, "/v1/incidents/2026-04-18-openai-down").Body, &body)

	if body.Data.Slug != "2026-04-18-openai-down" {
		t.Errorf("slug: got %q", body.Data.Slug)
	}
	if body.Data.ProviderID != "openai" {
		t.Errorf("provider_id: got %q", body.Data.ProviderID)
	}
	// nil slices must come back as [], not null
	if body.Data.AffectedModels == nil {
		t.Error("affected_models must not be null")
	}
	if body.Data.AffectedRegions == nil {
		t.Error("affected_regions must not be null")
	}
}

func TestGetIncident_ByUUID(t *testing.T) {
	inc := fixtureIncident("openai", "slug-uuid", "resolved", "minor")
	store := &fakeStore{incidents: []pgstore.Incident{inc}}
	srv := api.New(store)

	rec := doGet(t, srv, "/v1/incidents/"+inc.ID.String())
	if rec.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rec.Code)
	}
}

func TestGetIncident_NotFound(t *testing.T) {
	srv := api.New(&fakeStore{})
	rec := doGet(t, srv, "/v1/incidents/no-such-slug")
	if rec.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want 404", rec.Code)
	}
}

func TestListIncidents_LimitParam(t *testing.T) {
	incs := make([]pgstore.Incident, 10)
	for i := range incs {
		incs[i] = fixtureIncident("openai", "slug-"+string(rune('a'+i)), "resolved", "minor")
	}
	store := &fakeStore{incidents: incs}
	srv := api.New(store)

	var body struct {
		Data []any `json:"data"`
	}
	mustDecode(t, doGet(t, srv, "/v1/incidents?limit=3").Body, &body)

	if len(body.Data) != 3 {
		t.Errorf("data length: got %d, want 3", len(body.Data))
	}
}
