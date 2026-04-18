// Command prober loads active providers from Postgres, then schedules
// ProbeLightInference calls for every active model at a fixed interval.
//
// Required environment variables:
//
//	DATABASE_URL  Postgres DSN (pgx5:// scheme not needed here — plain postgres:// works with pgx)
//	REGION_ID     Identifier for this prober node, e.g. "us-west-2"
//
// Optional environment variables:
//
//	PROBE_INTERVAL    Duration between probe rounds (default: 60s)
//	PROBE_TIMEOUT     Per-probe deadline (default: 30s)
//	PROBE_CONCURRENCY Max simultaneous probes (default: 8)
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/llmstatus/llmstatus/internal/probes"
	"github.com/llmstatus/llmstatus/internal/probes/adapters"
	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

func main() {
	regionID := requireEnv("REGION_ID")
	dbURL := requireEnv("DATABASE_URL")

	interval := envDuration("PROBE_INTERVAL", 60*time.Second)
	timeout := envDuration("PROBE_TIMEOUT", 30*time.Second)
	concurrency := envInt("PROBE_CONCURRENCY", 8)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		slog.Error("prober: cannot connect to postgres", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	configs, err := loadProviderConfigs(ctx, pool)
	if err != nil {
		slog.Error("prober: cannot load providers", "err", err)
		os.Exit(1)
	}
	if len(configs) == 0 {
		slog.Warn("prober: no active providers with registered adapters — exiting")
		os.Exit(0)
	}

	r := probes.New(
		configs,
		probes.LogSink{},
		regionID,
		probes.WithInterval(interval),
		probes.WithProbeTimeout(timeout),
		probes.WithConcurrency(concurrency),
	)

	slog.Info("prober: starting",
		"region", regionID,
		"providers", len(configs),
		"interval", interval,
		"concurrency", concurrency,
	)

	if err := r.Run(ctx); err != nil {
		slog.Info("prober: stopped", "reason", err)
	}
}

func loadProviderConfigs(ctx context.Context, pool *pgxpool.Pool) ([]probes.ProviderConfig, error) {
	q := pgstore.New(pool)

	providers, err := q.ListActiveProviders(ctx)
	if err != nil {
		return nil, err
	}

	var configs []probes.ProviderConfig
	for _, p := range providers {
		adapter, ok := adapters.Get(p.ID)
		if !ok {
			slog.Warn("prober: no adapter registered, skipping", "provider", p.ID)
			continue
		}

		models, err := q.ListModelsByProvider(ctx, p.ID)
		if err != nil {
			return nil, err
		}

		modelIDs := make([]string, 0, len(models))
		for _, m := range models {
			modelIDs = append(modelIDs, m.ModelID)
		}

		if len(modelIDs) == 0 {
			slog.Warn("prober: provider has no models, skipping", "provider", p.ID)
			continue
		}

		configs = append(configs, probes.ProviderConfig{
			Adapter: adapter,
			Models:  modelIDs,
		})
	}
	return configs, nil
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		slog.Error("prober: required environment variable not set", "key", key)
		os.Exit(1)
	}
	return v
}

func envDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		slog.Warn("prober: invalid duration, using default", "key", key, "value", v, "default", def)
		return def
	}
	return d
}

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		slog.Warn("prober: invalid int, using default", "key", key, "value", v, "default", def)
		return def
	}
	return n
}
