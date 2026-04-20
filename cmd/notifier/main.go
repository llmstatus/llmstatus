// Command notifier polls for incident changes and delivers alerts via email and webhook.
//
// Required environment variables:
//
//	DATABASE_URL      Postgres DSN
//	RESEND_API_KEY    Resend email API key
//
// Optional:
//
//	EMAIL_FROM              Sender address (default "noreply@llmstatus.io")
//	SITE_URL                Public site URL (default "https://llmstatus.io")
//	NOTIFIER_POLL_INTERVAL  Poll cadence (default "30s")
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/llmstatus/llmstatus/internal/email"
	"github.com/llmstatus/llmstatus/internal/notifier"
	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

func main() {
	dbURL := requireEnv("DATABASE_URL")
	resendKey := requireEnv("RESEND_API_KEY")
	emailFrom := envOr("EMAIL_FROM", "noreply@llmstatus.io")
	siteURL := envOr("SITE_URL", "https://llmstatus.io")
	interval := envDuration("NOTIFIER_POLL_INTERVAL", 30*time.Second)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		slog.Error("notifier: cannot connect to postgres", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	n := notifier.New(notifier.Config{
		Store:    pgstore.New(pool),
		Email:    email.New(resendKey, emailFrom),
		SiteURL:  siteURL,
		Interval: interval,
	})

	slog.Info("notifier: starting", "interval", interval)
	n.Run(ctx)
	slog.Info("notifier: stopped")
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		slog.Error("notifier: required env var not set", "key", key)
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

func envDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
		slog.Warn("notifier: invalid duration, using default", "key", key, "default", def)
	}
	return def
}
