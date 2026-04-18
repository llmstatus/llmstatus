// Package influx provides a write-only client for InfluxDB 3.
//
// Writes use the InfluxDB v2 line-protocol HTTP endpoint
// (POST /api/v2/write), which InfluxDB 3 exposes for backward
// compatibility with no additional dependencies beyond net/http.
package influx

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/llmstatus/llmstatus/internal/probes"
)

// Writer persists probe results to InfluxDB as line-protocol points.
type Writer interface {
	WriteProbeResult(ctx context.Context, r probes.ProbeResult) error
	Close() error
}

// Config holds the connection parameters for the InfluxDB 3 write API.
type Config struct {
	// Host is the base URL, e.g. "https://us-east-1-1.aws.cloud2.influxdata.com".
	Host string
	// Token is the InfluxDB API token (Authorization: Token <token>).
	Token string
	// Database is the InfluxDB 3 database name (mapped to bucket in the v2 API).
	Database string
}

// NewWriter returns a Writer that POSTs line-protocol to the InfluxDB v2
// write endpoint. The caller must call Close when done (currently a no-op,
// reserved for connection pool teardown).
func NewWriter(cfg Config, client *http.Client) Writer {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &lineWriter{cfg: cfg, client: client}
}

type lineWriter struct {
	cfg    Config
	client *http.Client
}

func (w *lineWriter) WriteProbeResult(ctx context.Context, r probes.ProbeResult) error {
	line := toLineProtocol(r)
	url := w.cfg.Host + "/api/v2/write?precision=ns&bucket=" + w.cfg.Database

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(line))
	if err != nil {
		return fmt.Errorf("influx: build request: %w", err)
	}
	req.Header.Set("Authorization", "Token "+w.cfg.Token)
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("influx: write: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("influx: write status %d: %s", resp.StatusCode, bytes.TrimSpace(body))
	}
	return nil
}

func (w *lineWriter) Close() error { return nil }

// toLineProtocol converts a ProbeResult into a single InfluxDB line-protocol
// string. Tags are indexed; fields are queryable values; timestamp is nanoseconds.
func toLineProtocol(r probes.ProbeResult) string {
	// Tags (indexed, low-cardinality). Values must escape commas, spaces, equals.
	tags := "provider_id=" + escapeTag(r.ProviderID) +
		",model=" + escapeTag(r.Model) +
		",probe_type=" + escapeTag(r.ProbeType) +
		",region_id=" + escapeTag(r.RegionID)
	if r.ErrorClass != "" {
		tags += ",error_class=" + escapeTag(string(r.ErrorClass))
	}

	// Fields (queryable values).
	success := "false"
	if r.Success {
		success = "true"
	}
	fields := fmt.Sprintf("success=%s,duration_ms=%di,http_status=%di",
		success, r.DurationMs, r.HTTPStatus)
	if r.TokensIn > 0 {
		fields += fmt.Sprintf(",tokens_in=%di", r.TokensIn)
	}
	if r.TokensOut > 0 {
		fields += fmt.Sprintf(",tokens_out=%di", r.TokensOut)
	}

	ts := r.StartedAt.UnixNano()
	return fmt.Sprintf("probes,%s %s %d", tags, fields, ts)
}

// escapeTag escapes special characters in InfluxDB line-protocol tag keys
// and values: commas, spaces, and equals signs require a backslash prefix.
func escapeTag(s string) string {
	s = strings.ReplaceAll(s, ",", `\,`)
	s = strings.ReplaceAll(s, " ", `\ `)
	s = strings.ReplaceAll(s, "=", `\=`)
	return s
}
