// Command api serves the public read API for llmstatus.io.
//
// Required environment variables:
//
//	DATABASE_URL      Postgres DSN (pgx connection string)
//	INFLUX_HOST       InfluxDB 3 base URL
//	INFLUX_TOKEN      InfluxDB auth token
//	INFLUX_DATABASE   InfluxDB database / bucket name
//
// Optional:
//
//	API_ADDR  Listen address (default ":8081")
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/llmstatus/llmstatus/internal/api"
	"github.com/llmstatus/llmstatus/internal/store/influx"
	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

func main() {
	dbURL := requireEnv("DATABASE_URL")
	influxHost := requireEnv("INFLUX_HOST")
	influxToken := requireEnv("INFLUX_TOKEN")
	influxDB := requireEnv("INFLUX_DATABASE")
	addr := envOr("API_ADDR", ":8081")
	rateLimit := envInt("API_RATE_LIMIT", 60)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		slog.Error("api: cannot connect to postgres", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	store := pgstore.New(pool)
	historyReader := influx.NewHistoryReader(influx.HistoryReaderConfig{
		Host:     influxHost,
		Token:    influxToken,
		Database: influxDB,
	}, nil)
	limiter := api.NewRateLimiter(rateLimit, time.Minute)

	srv := &http.Server{
		Addr: addr,
		Handler: api.New(store,
			api.WithHistoryReader(historyReader),
			api.WithLiveStatsReader(historyReader),
			api.WithRateLimiter(limiter),
		),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("api: listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("api: server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("api: shutting down")

	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		slog.Error("api: shutdown error", "err", err)
	}
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		slog.Error("api: required environment variable not set", "key", key)
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

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
		slog.Warn("api: invalid env var, using default", "key", key, "default", def)
	}
	return def
}
