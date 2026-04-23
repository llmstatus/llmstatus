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
//	API_ADDR               Listen address (default ":8081")
//	JWT_SECRET             Enable auth routes; signs session tokens
//	RESEND_API_KEY         Resend email API key (required when JWT_SECRET set)
//	EMAIL_FROM             Sender address (default "noreply@llmstatus.io")
//	INTERNAL_SECRET        Shared secret for /auth/oauth/upsert (required when JWT_SECRET set)
//	REDIS_URL              Redis DSN — enables per-email OTP rate limiting when set
//	SITE_URL               Public site URL (default "https://llmstatus.io")
//	GOOGLE_CLIENT_ID       Google OAuth client ID
//	GOOGLE_CLIENT_SECRET   Google OAuth client secret
//	GITHUB_CLIENT_ID       GitHub OAuth client ID
//	GITHUB_CLIENT_SECRET   GitHub OAuth client secret
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

	"github.com/redis/go-redis/v9"

	"github.com/llmstatus/llmstatus/internal/api"
	"github.com/llmstatus/llmstatus/internal/email"
	"github.com/llmstatus/llmstatus/internal/keyenc"
	"github.com/llmstatus/llmstatus/internal/otprl"
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

	opts := []func(*api.Server){
		api.WithPinger(pool),
		api.WithHistoryReader(historyReader),
		api.WithLiveStatsReader(historyReader),
		api.WithRateLimiter(limiter),
	}

	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		emailFrom := envOr("EMAIL_FROM", "noreply@llmstatus.io")
		authCfg := &api.AuthConfig{
			Store:          store,
			Email:          email.New(requireEnv("RESEND_API_KEY"), emailFrom),
			JWTSecret:      jwtSecret,
			InternalSecret: requireEnv("INTERNAL_SECRET"),
		}
		if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
			opt, err := redis.ParseURL(redisURL)
			if err != nil {
				slog.Error("api: invalid REDIS_URL", "err", err)
				os.Exit(1)
			}
			authCfg.OTPLimiter = otprl.NewRedis(redis.NewClient(opt))
			slog.Info("api: OTP rate-limiting enabled")
		}
		opts = append(opts, api.WithAuth(authCfg))
		slog.Info("api: auth enabled")
	}

	if hexKey := os.Getenv("SPONSOR_KEY_ENCRYPTION_KEY"); hexKey != "" {
		enc, err := keyenc.New(hexKey)
		if err != nil {
			slog.Error("api: invalid SPONSOR_KEY_ENCRYPTION_KEY", "err", err)
			os.Exit(1)
		}
		opts = append(opts, api.WithKeyEncrypter(enc))
		slog.Info("api: sponsor key encryption enabled")
	}

	srv := &http.Server{
		Addr:    addr,
		Handler: api.New(store, opts...),
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

	// Gracefully shutdown WebSocket hub
	hubShutCtx, hubCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer hubCancel()
	if err := api.GetGlobalHub().Shutdown(hubShutCtx); err != nil {
		slog.Error("api: hub shutdown error", "err", err)
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
