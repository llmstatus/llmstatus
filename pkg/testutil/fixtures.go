package testutil

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

// FixtureProvider inserts a minimal Provider row and returns its ID.
// Safe to call from parallel tests — each invocation uses a unique ID suffix.
func FixtureProvider(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()

	id := "fixture_" + t.Name()
	q := pgstore.New(pool)
	err := q.UpsertProvider(context.Background(), pgstore.UpsertProviderParams{
		ID:         id,
		Name:       "Fixture Provider",
		Category:   "official",
		BaseUrl:    "https://api.fixture.test",
		AuthType:   "bearer",
		Region:     "global",
		Active:     true,
		Config:     json.RawMessage(`{}`),
		ProbeScope: "global",
	})
	if err != nil {
		t.Fatalf("FixtureProvider: upsert provider %q: %v", id, err)
	}
	return id
}

// FixtureModel inserts a minimal Model row for the given provider and returns
// the generated model row ID.
func FixtureModel(t *testing.T, pool *pgxpool.Pool, providerID string) int64 {
	t.Helper()

	q := pgstore.New(pool)
	m, err := q.UpsertModel(context.Background(), pgstore.UpsertModelParams{
		ProviderID:  providerID,
		ModelID:     "fixture-model",
		DisplayName: "Fixture Model",
		ModelType:   "chat",
		Active:      true,
	})
	if err != nil {
		t.Fatalf("FixtureModel: upsert model for provider %q: %v", providerID, err)
	}
	return m.ID
}
