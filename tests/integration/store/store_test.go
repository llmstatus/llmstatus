package store_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/goleak"

	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
	"github.com/llmstatus/llmstatus/pkg/testutil"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

// TestIntegration_Provider covers UpsertProvider, GetProvider,
// ListProviders, and ListActiveProviders against a real Postgres schema.
func TestIntegration_Provider(t *testing.T) {
	pool := testutil.NewPostgres(t)
	ctx := context.Background()
	q := pgstore.New(pool)

	t.Run("upsert and get", func(t *testing.T) {
		err := q.UpsertProvider(ctx, pgstore.UpsertProviderParams{
			ID:         "test_openai",
			Name:       "OpenAI",
			Category:   "official",
			BaseUrl:    "https://api.openai.com",
			AuthType:   "bearer",
			Region:     "global",
			Active:     true,
			Config:     json.RawMessage(`{"models":["gpt-4o-mini"]}`),
			ProbeScope: "global",
		})
		if err != nil {
			t.Fatalf("UpsertProvider: %v", err)
		}

		got, err := q.GetProvider(ctx, "test_openai")
		if err != nil {
			t.Fatalf("GetProvider: %v", err)
		}
		if got.Name != "OpenAI" {
			t.Errorf("Name: got %q, want %q", got.Name, "OpenAI")
		}
		if got.Category != "official" {
			t.Errorf("Category: got %q, want %q", got.Category, "official")
		}
		if !got.Active {
			t.Error("Active: got false, want true")
		}
	})

	t.Run("upsert is idempotent", func(t *testing.T) {
		params := pgstore.UpsertProviderParams{
			ID:         "test_openai",
			Name:       "OpenAI (updated)",
			Category:   "official",
			BaseUrl:    "https://api.openai.com",
			AuthType:   "bearer",
			Region:     "us",
			Active:     true,
			Config:     json.RawMessage(`{}`),
			ProbeScope: "global",
		}
		if err := q.UpsertProvider(ctx, params); err != nil {
			t.Fatalf("UpsertProvider (2nd): %v", err)
		}
		got, err := q.GetProvider(ctx, "test_openai")
		if err != nil {
			t.Fatalf("GetProvider after upsert: %v", err)
		}
		if got.Name != "OpenAI (updated)" {
			t.Errorf("Name after upsert: got %q, want %q", got.Name, "OpenAI (updated)")
		}
		if got.Region != "us" {
			t.Errorf("Region after upsert: got %q, want %q", got.Region, "us")
		}
	})

	t.Run("list active only", func(t *testing.T) {
		_ = q.UpsertProvider(ctx, pgstore.UpsertProviderParams{
			ID: "test_inactive", Name: "Inactive", Category: "official",
			BaseUrl: "https://x.test", AuthType: "bearer", Region: "global",
			Active: false, Config: json.RawMessage(`{}`), ProbeScope: "global",
		})

		active, err := q.ListActiveProviders(ctx)
		if err != nil {
			t.Fatalf("ListActiveProviders: %v", err)
		}
		for _, p := range active {
			if !p.Active {
				t.Errorf("ListActiveProviders returned inactive provider %q", p.ID)
			}
		}
	})

	t.Run("set active flag", func(t *testing.T) {
		if err := q.SetProviderActive(ctx, pgstore.SetProviderActiveParams{
			ID: "test_openai", Active: false,
		}); err != nil {
			t.Fatalf("SetProviderActive: %v", err)
		}
		got, err := q.GetProvider(ctx, "test_openai")
		if err != nil {
			t.Fatalf("GetProvider after deactivate: %v", err)
		}
		if got.Active {
			t.Error("Expected provider to be inactive after SetProviderActive(false)")
		}
	})
}

// TestIntegration_Model covers UpsertModel and ListModelsByProvider.
func TestIntegration_Model(t *testing.T) {
	pool := testutil.NewPostgres(t)
	ctx := context.Background()
	q := pgstore.New(pool)

	providerID := testutil.FixtureProvider(t, pool)

	t.Run("upsert and get", func(t *testing.T) {
		m, err := q.UpsertModel(ctx, pgstore.UpsertModelParams{
			ProviderID:  providerID,
			ModelID:     "gpt-4o-mini",
			DisplayName: "GPT-4o mini",
			ModelType:   "chat",
			Active:      true,
		})
		if err != nil {
			t.Fatalf("UpsertModel: %v", err)
		}
		if m.ID == 0 {
			t.Error("UpsertModel returned zero ID")
		}
		if m.ModelID != "gpt-4o-mini" {
			t.Errorf("ModelID: got %q, want %q", m.ModelID, "gpt-4o-mini")
		}
	})

	t.Run("upsert updates display_name", func(t *testing.T) {
		m, err := q.UpsertModel(ctx, pgstore.UpsertModelParams{
			ProviderID:  providerID,
			ModelID:     "gpt-4o-mini",
			DisplayName: "GPT-4o mini (renamed)",
			ModelType:   "chat",
			Active:      true,
		})
		if err != nil {
			t.Fatalf("UpsertModel (2nd): %v", err)
		}
		if m.DisplayName != "GPT-4o mini (renamed)" {
			t.Errorf("DisplayName: got %q, want %q", m.DisplayName, "GPT-4o mini (renamed)")
		}
	})

	t.Run("list by provider", func(t *testing.T) {
		_, _ = q.UpsertModel(ctx, pgstore.UpsertModelParams{
			ProviderID: providerID, ModelID: "gpt-4o",
			DisplayName: "GPT-4o", ModelType: "chat", Active: true,
		})

		models, err := q.ListModelsByProvider(ctx, providerID)
		if err != nil {
			t.Fatalf("ListModelsByProvider: %v", err)
		}
		if len(models) < 2 {
			t.Errorf("ListModelsByProvider: got %d models, want ≥2", len(models))
		}
		for _, m := range models {
			if m.ProviderID != providerID {
				t.Errorf("ListModelsByProvider returned model for wrong provider %q", m.ProviderID)
			}
		}
	})

	t.Run("get specific model", func(t *testing.T) {
		got, err := q.GetModel(ctx, pgstore.GetModelParams{
			ProviderID: providerID,
			ModelID:    "gpt-4o-mini",
		})
		if err != nil {
			t.Fatalf("GetModel: %v", err)
		}
		if got.ProviderID != providerID {
			t.Errorf("ProviderID: got %q, want %q", got.ProviderID, providerID)
		}
	})
}

// TestIntegration_Incident covers CreateIncident, GetIncidentBySlug,
// GetOngoingByProviderAndRule, UpdateIncidentStatus, and ResolveIncident.
func TestIntegration_Incident(t *testing.T) {
	pool := testutil.NewPostgres(t)
	ctx := context.Background()
	q := pgstore.New(pool)

	providerID := testutil.FixtureProvider(t, pool)

	newIncident := func(slug, rule string) pgstore.CreateIncidentParams {
		return pgstore.CreateIncidentParams{
			Slug:            slug,
			ProviderID:      providerID,
			Severity:        "major",
			Title:           "Test incident: " + slug,
			Status:          "ongoing",
			AffectedModels:  []string{"gpt-4o-mini"},
			AffectedRegions: []string{"us-west-2"},
			StartedAt:       pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
			DetectionMethod: "auto",
			DetectionRule:   pgtype.Text{String: rule, Valid: true},
			MetricsSnapshot: json.RawMessage(`{"error_rate":0.15}`),
		}
	}

	t.Run("create and get by slug", func(t *testing.T) {
		inc, err := q.CreateIncident(ctx, newIncident("2026-test-openai-outage", "elevated_errors"))
		if err != nil {
			t.Fatalf("CreateIncident: %v", err)
		}
		if inc.Status != "ongoing" {
			t.Errorf("Status: got %q, want %q", inc.Status, "ongoing")
		}

		got, err := q.GetIncidentBySlug(ctx, "2026-test-openai-outage")
		if err != nil {
			t.Fatalf("GetIncidentBySlug: %v", err)
		}
		if !cmp.Equal(inc.ID, got.ID) {
			t.Errorf("ID mismatch: created %v, fetched %v", inc.ID, got.ID)
		}
	})

	t.Run("deduplication: get ongoing by provider and rule", func(t *testing.T) {
		_, _ = q.CreateIncident(ctx, newIncident("2026-test-dedup-incident", "provider_down"))

		ongoing, err := q.GetOngoingByProviderAndRule(ctx, pgstore.GetOngoingByProviderAndRuleParams{
			ProviderID:    providerID,
			DetectionRule: pgtype.Text{String: "provider_down", Valid: true},
		})
		if err != nil {
			t.Fatalf("GetOngoingByProviderAndRule: %v", err)
		}
		if ongoing.Slug != "2026-test-dedup-incident" {
			t.Errorf("Slug: got %q, want %q", ongoing.Slug, "2026-test-dedup-incident")
		}
	})

	t.Run("update status and resolve", func(t *testing.T) {
		inc, err := q.CreateIncident(ctx, newIncident("2026-test-resolve", "elevated_latency"))
		if err != nil {
			t.Fatalf("CreateIncident: %v", err)
		}

		if err := q.UpdateIncidentStatus(ctx, pgstore.UpdateIncidentStatusParams{
			ID:     inc.ID,
			Status: "monitoring",
		}); err != nil {
			t.Fatalf("UpdateIncidentStatus: %v", err)
		}

		resolvedAt := pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
		if err := q.ResolveIncident(ctx, pgstore.ResolveIncidentParams{
			ID:         inc.ID,
			ResolvedAt: resolvedAt,
		}); err != nil {
			t.Fatalf("ResolveIncident: %v", err)
		}

		got, err := q.GetIncidentByID(ctx, inc.ID)
		if err != nil {
			t.Fatalf("GetIncidentByID after resolve: %v", err)
		}
		if got.Status != "resolved" {
			t.Errorf("Status after resolve: got %q, want %q", got.Status, "resolved")
		}
		if !got.ResolvedAt.Valid {
			t.Error("ResolvedAt should be set after ResolveIncident")
		}
	})

	t.Run("list incidents", func(t *testing.T) {
		incidents, err := q.ListIncidents(ctx, pgstore.ListIncidentsParams{
			Limit: 10, Offset: 0,
		})
		if err != nil {
			t.Fatalf("ListIncidents: %v", err)
		}
		if len(incidents) == 0 {
			t.Error("ListIncidents returned empty slice (expected rows from other subtests)")
		}
	})
}
