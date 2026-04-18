// Command ingest is an HTTP server that receives probe results from probers
// and writes them to InfluxDB 3 as line-protocol points.
//
// Required environment variables:
//
//	INFLUX_HOST      InfluxDB 3 host URL, e.g. "https://us-east-1-1.aws.cloud2.influxdata.com"
//	INFLUX_TOKEN     InfluxDB API token
//	INFLUX_DATABASE  Database name (InfluxDB 3) or bucket (v2 API)
//
// Optional:
//
//	INGEST_ADDR  Listen address (default ":8080")
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/llmstatus/llmstatus/internal/ingest"
	"github.com/llmstatus/llmstatus/internal/store/influx"
)

func main() {
	influxHost := requireEnv("INFLUX_HOST")
	influxToken := requireEnv("INFLUX_TOKEN")
	influxDB := requireEnv("INFLUX_DATABASE")
	addr := envOr("INGEST_ADDR", ":8080")

	writer := influx.NewWriter(influx.Config{
		Host:     influxHost,
		Token:    influxToken,
		Database: influxDB,
	}, nil)
	defer func() { _ = writer.Close() }()

	mux := http.NewServeMux()
	mux.Handle("/v1/probe", ingest.NewHandler(writer))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	go func() {
		slog.Info("ingest: listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("ingest: server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("ingest: shutting down")

	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		slog.Error("ingest: shutdown error", "err", err)
	}
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		slog.Error("ingest: required environment variable not set", "key", key)
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
