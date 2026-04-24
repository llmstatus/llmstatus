package api

import (
	"context"

	"github.com/google/uuid"

	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

// Store is the subset of pgstore.Querier used by the read API.
// pgstore.Queries satisfies this interface at compile time (see server.go).
type Store interface {
	ListActiveProviders(ctx context.Context) ([]pgstore.Provider, error)
	ListProvidersForScope(ctx context.Context, probeScope string) ([]pgstore.Provider, error)
	GetProvider(ctx context.Context, id string) (pgstore.Provider, error)
	ListModelsByProvider(ctx context.Context, providerID string) ([]pgstore.Model, error)
	ListIncidents(ctx context.Context, arg pgstore.ListIncidentsParams) ([]pgstore.Incident, error)
	ListIncidentsByProvider(ctx context.Context, arg pgstore.ListIncidentsByProviderParams) ([]pgstore.Incident, error)
	ListIncidentsByStatus(ctx context.Context, arg pgstore.ListIncidentsByStatusParams) ([]pgstore.Incident, error)
	GetIncidentByID(ctx context.Context, id uuid.UUID) (pgstore.Incident, error)
	GetIncidentBySlug(ctx context.Context, slug string) (pgstore.Incident, error)
	InsertUserReport(ctx context.Context, arg pgstore.InsertUserReportParams) error
	UserReportHistogram(ctx context.Context, providerID string) ([]pgstore.UserReportHistogramRow, error)

	// Sponsors
	ListActiveSponsors(ctx context.Context) ([]pgstore.Sponsor, error)
	CreateSponsor(ctx context.Context, arg pgstore.CreateSponsorParams) (pgstore.Sponsor, error)
	GetSponsorByUserID(ctx context.Context, userID int64) (pgstore.Sponsor, error)
	UpdateSponsor(ctx context.Context, arg pgstore.UpdateSponsorParams) (pgstore.Sponsor, error)
	ListSponsorKeys(ctx context.Context, sponsorID string) ([]pgstore.SponsorKey, error)
	UpsertSponsorKey(ctx context.Context, arg pgstore.UpsertSponsorKeyParams) (pgstore.SponsorKey, error)
	DeleteSponsorKey(ctx context.Context, arg pgstore.DeleteSponsorKeyParams) error
}
