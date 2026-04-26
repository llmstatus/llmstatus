package api_test

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/llmstatus/llmstatus/internal/store/influx"
	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

// fakeLiveStatsReader implements api.LiveStatsReader for unit tests.
type fakeLiveStatsReader struct {
	stats []influx.ProviderLiveStat
	err   error
}

func (f *fakeLiveStatsReader) AllProviderLiveStats(_ context.Context) ([]influx.ProviderLiveStat, error) {
	return f.stats, f.err
}

func (f *fakeLiveStatsReader) AllModelLiveStats(_ context.Context) ([]influx.ModelLiveStat, error) {
	return nil, nil
}

func (f *fakeLiveStatsReader) AllModelSparklines(_ context.Context) (map[string][]float64, error) {
	return nil, nil
}

func (f *fakeLiveStatsReader) ProviderRegionStats(_ context.Context, _ string) ([]influx.RegionLiveStat, error) {
	return nil, nil
}

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

func (f *fakeStore) InsertUserReport(_ context.Context, _ pgstore.InsertUserReportParams) error {
	return f.err
}

func (f *fakeStore) UserReportHistogram(_ context.Context, _ string) ([]pgstore.UserReportHistogramRow, error) {
	return nil, f.err
}

func (f *fakeStore) ListActiveSponsors(_ context.Context) ([]pgstore.Sponsor, error) {
	return nil, f.err
}

func (f *fakeStore) ListPendingSponsors(_ context.Context) ([]pgstore.Sponsor, error) {
	return nil, f.err
}

func (f *fakeStore) CreateSponsor(_ context.Context, arg pgstore.CreateSponsorParams) (pgstore.Sponsor, error) {
	return pgstore.Sponsor{}, f.err
}

func (f *fakeStore) GetSponsorByID(_ context.Context, _ string) (pgstore.Sponsor, error) {
	return pgstore.Sponsor{}, pgx.ErrNoRows
}

func (f *fakeStore) GetSponsorByUserID(_ context.Context, userID pgtype.Int8) (pgstore.Sponsor, error) {
	return pgstore.Sponsor{}, pgx.ErrNoRows
}

func (f *fakeStore) UpdateSponsor(_ context.Context, arg pgstore.UpdateSponsorParams) (pgstore.Sponsor, error) {
	return pgstore.Sponsor{}, f.err
}

func (f *fakeStore) ApproveSponsor(_ context.Context, _ string) (pgstore.Sponsor, error) {
	return pgstore.Sponsor{}, f.err
}

func (f *fakeStore) RejectSponsor(_ context.Context, _ string) (pgstore.Sponsor, error) {
	return pgstore.Sponsor{}, f.err
}

func (f *fakeStore) ListSponsorKeys(_ context.Context, sponsorID string) ([]pgstore.SponsorKey, error) {
	return nil, f.err
}

func (f *fakeStore) UpsertSponsorKey(_ context.Context, arg pgstore.UpsertSponsorKeyParams) (pgstore.SponsorKey, error) {
	return pgstore.SponsorKey{}, f.err
}

func (f *fakeStore) DeleteSponsorKey(_ context.Context, arg pgstore.DeleteSponsorKeyParams) error {
	return f.err
}

func (f *fakeStore) ListProvidersForScope(_ context.Context, probeScope string) ([]pgstore.Provider, error) {
	return f.providers, f.err
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
