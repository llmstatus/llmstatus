package detector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ProbeStats holds aggregated probe counts for one provider over a time window.
type ProbeStats struct {
	ProviderID string
	Total      int64
	Errors     int64
}

// ErrorRate returns the fraction of failed probes, or 0 when Total is 0.
func (s ProbeStats) ErrorRate() float64 {
	if s.Total == 0 {
		return 0
	}
	return float64(s.Errors) / float64(s.Total)
}

// ProbeReader fetches aggregated probe statistics from the time-series store.
type ProbeReader interface {
	// ErrorRateByProvider returns per-provider probe stats for the given window.
	ErrorRateByProvider(ctx context.Context, window time.Duration) ([]ProbeStats, error)
}

// InfluxReaderConfig holds connection parameters for the InfluxDB 3 query API.
type InfluxReaderConfig struct {
	Host     string // e.g. "https://us-east-1-1.aws.cloud2.influxdata.com"
	Token    string
	Database string
}

// NewInfluxReader returns a ProbeReader that queries InfluxDB 3 via the
// HTTP SQL endpoint (POST /api/v3/query_sql), which requires no gRPC dep.
func NewInfluxReader(cfg InfluxReaderConfig, client *http.Client) ProbeReader {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	return &influxReader{cfg: cfg, client: client}
}

type influxReader struct {
	cfg    InfluxReaderConfig
	client *http.Client
}

func (r *influxReader) ErrorRateByProvider(ctx context.Context, window time.Duration) ([]ProbeStats, error) {
	sql := fmt.Sprintf(
		`SELECT provider_id,
		        COUNT(*) AS total,
		        COUNT(*) FILTER (WHERE success = false) AS errors
		 FROM probes
		 WHERE time >= now() - INTERVAL '%d seconds'
		 GROUP BY provider_id`,
		int(window.Seconds()),
	)

	body, err := json.Marshal(map[string]string{
		"q":      sql,
		"db":     r.cfg.Database,
		"format": "json",
	})
	if err != nil {
		return nil, fmt.Errorf("detector: marshal query: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		r.cfg.Host+"/api/v3/query_sql", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("detector: build request: %w", err)
	}
	req.Header.Set("Authorization", "Token "+r.cfg.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("detector: query influx: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("detector: influx status %d: %s", resp.StatusCode, detail)
	}

	// InfluxDB 3 returns a JSON array of objects with format=json.
	var rows []struct {
		ProviderID string  `json:"provider_id"`
		Total      float64 `json:"total"` // JSON numbers decode as float64
		Errors     float64 `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, fmt.Errorf("detector: decode response: %w", err)
	}

	stats := make([]ProbeStats, 0, len(rows))
	for _, row := range rows {
		if row.ProviderID == "" {
			continue
		}
		stats = append(stats, ProbeStats{
			ProviderID: row.ProviderID,
			Total:      int64(row.Total),
			Errors:     int64(row.Errors),
		})
	}
	return stats, nil
}
