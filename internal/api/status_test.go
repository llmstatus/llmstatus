package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/llmstatus/llmstatus/internal/api"
	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

// fakePinger satisfies api.Pinger for healthz tests.
type fakePinger struct{ err error }

func (f *fakePinger) Ping(_ context.Context) error { return f.err }

// ---- /v1/status tests -------------------------------------------------------

func TestGetStatus_AllOperational(t *testing.T) {
	store := &fakeStore{
		providers: []pgstore.Provider{
			fixtureProvider("openai", "OpenAI"),
			fixtureProvider("anthropic", "Anthropic"),
		},
	}
	srv := api.New(store)
	rec := doGet(t, srv, "/v1/status")

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", rec.Code)
	}
	var body struct {
		Data struct {
			Status string `json:"status"`
			Counts struct {
				Total       int `json:"total"`
				Operational int `json:"operational"`
				Degraded    int `json:"degraded"`
				Down        int `json:"down"`
			} `json:"counts"`
		} `json:"data"`
	}
	mustDecode(t, rec.Body, &body)

	if body.Data.Status != "operational" {
		t.Errorf("status: got %q, want operational", body.Data.Status)
	}
	if body.Data.Counts.Total != 2 {
		t.Errorf("total: got %d, want 2", body.Data.Counts.Total)
	}
	if body.Data.Counts.Operational != 2 {
		t.Errorf("operational: got %d, want 2", body.Data.Counts.Operational)
	}
}

func TestGetStatus_OneDown(t *testing.T) {
	inc := fixtureIncident("openai", "2026-04-18-openai-down", "ongoing", "critical")
	store := &fakeStore{
		providers: []pgstore.Provider{
			fixtureProvider("openai", "OpenAI"),
			fixtureProvider("anthropic", "Anthropic"),
		},
		incidents: []pgstore.Incident{inc},
	}
	srv := api.New(store)
	rec := doGet(t, srv, "/v1/status")

	var body struct {
		Data struct {
			Status string `json:"status"`
			Counts struct {
				Total       int `json:"total"`
				Operational int `json:"operational"`
				Down        int `json:"down"`
			} `json:"counts"`
		} `json:"data"`
	}
	mustDecode(t, rec.Body, &body)

	if body.Data.Status != "down" {
		t.Errorf("status: got %q, want down", body.Data.Status)
	}
	if body.Data.Counts.Down != 1 {
		t.Errorf("down: got %d, want 1", body.Data.Counts.Down)
	}
	if body.Data.Counts.Operational != 1 {
		t.Errorf("operational: got %d, want 1", body.Data.Counts.Operational)
	}
}

func TestGetStatus_OneDegraded_NotDown(t *testing.T) {
	inc := fixtureIncident("anthropic", "2026-04-18-anthropic-deg", "ongoing", "major")
	store := &fakeStore{
		providers: []pgstore.Provider{
			fixtureProvider("openai", "OpenAI"),
			fixtureProvider("anthropic", "Anthropic"),
		},
		incidents: []pgstore.Incident{inc},
	}
	srv := api.New(store)
	rec := doGet(t, srv, "/v1/status")

	var body struct {
		Data struct {
			Status string `json:"status"`
			Counts struct {
				Degraded int `json:"degraded"`
			} `json:"counts"`
		} `json:"data"`
	}
	mustDecode(t, rec.Body, &body)

	if body.Data.Status != "degraded" {
		t.Errorf("status: got %q, want degraded", body.Data.Status)
	}
	if body.Data.Counts.Degraded != 1 {
		t.Errorf("degraded: got %d, want 1", body.Data.Counts.Degraded)
	}
}

func TestGetStatus_Empty(t *testing.T) {
	srv := api.New(&fakeStore{})
	rec := doGet(t, srv, "/v1/status")

	var body struct {
		Data json.RawMessage `json:"data"`
	}
	mustDecode(t, rec.Body, &body)
	// Should return operational when no providers configured.
	var d struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(body.Data, &d); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if d.Status != "operational" {
		t.Errorf("status: got %q, want operational", d.Status)
	}
}

// ---- /healthz tests ---------------------------------------------------------

func TestHealthz_NoPinger_Always200(t *testing.T) {
	srv := api.New(&fakeStore{}) // no WithPinger
	rec := doGet(t, srv, "/healthz")
	if rec.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rec.Code)
	}
}

func TestHealthz_PingerOK_Returns200(t *testing.T) {
	srv := api.New(&fakeStore{}, api.WithPinger(&fakePinger{}))
	rec := doGet(t, srv, "/healthz")
	if rec.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rec.Code)
	}
}

func TestHealthz_PingerFail_Returns503(t *testing.T) {
	srv := api.New(&fakeStore{}, api.WithPinger(&fakePinger{err: errors.New("connection refused")}))
	rec := doGet(t, srv, "/healthz")
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status: got %d, want 503", rec.Code)
	}
}
