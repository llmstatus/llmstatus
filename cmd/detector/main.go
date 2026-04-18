// Command detector evaluates detection rules every interval and manages
// incidents in Postgres automatically.
//
// Required environment variables:
//
//	DATABASE_URL      Postgres DSN (pgx connection string)
//	INFLUX_HOST       InfluxDB 3 base URL (e.g. "https://us-east-1-1.aws.cloud2.influxdata.com")
//	INFLUX_TOKEN      InfluxDB auth token
//	INFLUX_DATABASE   InfluxDB database / bucket name
//
// Optional:
//
//	DETECTOR_INTERVAL  Evaluation interval (default "60s")
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/llmstatus/llmstatus/internal/detector"
	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

func main() {
	dbURL := requireEnv("DATABASE_URL")
	influxHost := requireEnv("INFLUX_HOST")
	influxToken := requireEnv("INFLUX_TOKEN")
	influxDB := requireEnv("INFLUX_DATABASE")
	interval := parseDuration(envOr("DETECTOR_INTERVAL", "60s"))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		slog.Error("detector: cannot connect to postgres", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	reader := detector.NewInfluxReader(detector.InfluxReaderConfig{
		Host:     influxHost,
		Token:    influxToken,
		Database: influxDB,
	}, nil)

	store := pgstore.New(pool)
	runner := detector.New(reader, store, interval)

	slog.Info("detector: starting", "interval", interval)
	if err := runner.Run(ctx); err != nil {
		slog.Info("detector: stopped", "reason", err)
	}
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		slog.Error("detector: required environment variable not set", "key", key)
		os.Exit(1)
	}
	return v
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		slog.Error("detector: invalid DETECTOR_INTERVAL", "value", s, "err", err)
		os.Exit(1)
	}
	return d
}
