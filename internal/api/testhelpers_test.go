package api_test

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

// fakeStore implements api.Store for unit tests.
type fakeStore struct {
	providers []pgstore.Provider
	models    []pgstore.Model
	incidents []pgstore.Incident
	err       error
}

func (f *fakeStore) ListActiveProviders(_ context.Context) ([]pgstore.Provider, error) {
	return f.providers, f.err
}

func (f *fakeStore) GetProvider(_ context.Context, id string) (pgstore.Provider, error) {
	for _, p := range f.providers {
		if p.ID == id {
			return p, nil
		}
	}
	return pgstore.Provider{}, pgx.ErrNoRows
}

func (f *fakeStore) ListModelsByProvider(_ context.Context, providerID string) ([]pgstore.Model, error) {
	var out []pgstore.Model
	for _, m := range f.models {
		if m.ProviderID == providerID {
			out = append(out, m)
		}
	}
	return out, f.err
}

func (f *fakeStore) ListIncidents(_ context.Context, arg pgstore.ListIncidentsParams) ([]pgstore.Incident, error) {
	end := int(arg.Offset) + int(arg.Limit)
	if end > len(f.incidents) {
		end = len(f.incidents)
	}
	if int(arg.Offset) >= len(f.incidents) {
		return []pgstore.Incident{}, f.err
	}
	return f.incidents[arg.Offset:end], f.err
}

func (f *fakeStore) ListIncidentsByProvider(_ context.Context, arg pgstore.ListIncidentsByProviderParams) ([]pgstore.Incident, error) {
	var out []pgstore.Incident
	for _, inc := range f.incidents {
		if inc.ProviderID == arg.ProviderID {
			out = append(out, inc)
		}
	}
	return out, f.err
}

func (f *fakeStore) ListIncidentsByStatus(_ context.Context, arg pgstore.ListIncidentsByStatusParams) ([]pgstore.Incident, error) {
	var out []pgstore.Incident
	for _, inc := range f.incidents {
		if inc.Status == arg.Status {
			out = append(out, inc)
		}
	}
	return out, f.err
}

func (f *fakeStore) GetIncidentByID(_ context.Context, id uuid.UUID) (pgstore.Incident, error) {
	for _, inc := range f.incidents {
		if inc.ID == id {
			return inc, nil
		}
	}
	return pgstore.Incident{}, pgx.ErrNoRows
}

func (f *fakeStore) GetIncidentBySlug(_ context.Context, slug string) (pgstore.Incident, error) {
	for _, inc := range f.incidents {
		if inc.Slug == slug {
			return inc, nil
		}
	}
	return pgstore.Incident{}, pgx.ErrNoRows
}

// ---- fixtures ---------------------------------------------------------------

func fixtureProvider(id, name string) pgstore.Provider {
	return pgstore.Provider{
		ID:       id,
		Name:     name,
		Category: "official",
		Region:   "global",
		Active:   true,
	}
}

func fixtureModel(providerID, modelID string) pgstore.Model {
	return pgstore.Model{
		ID:          1,
		ProviderID:  providerID,
		ModelID:     modelID,
		DisplayName: modelID,
		ModelType:   "chat",
		Active:      true,
	}
}

func fixtureIncident(providerID, slug, status, severity string) pgstore.Incident {
	return pgstore.Incident{
		ID:         uuid.New(),
		Slug:       slug,
		ProviderID: providerID,
		Severity:   severity,
		Title:      "Test incident",
		Status:     status,
		StartedAt: pgtype.Timestamptz{
			Time:  time.Now().UTC(),
			Valid: true,
		},
		DetectionMethod: "auto",
	}
}
