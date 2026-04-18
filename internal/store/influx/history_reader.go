package influx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HistoryBucket holds aggregated probe stats for one time bucket.
type HistoryBucket struct {
	Timestamp time.Time `json:"timestamp"`
	Total     int64     `json:"total"`
	Errors    int64     `json:"errors"`
	Uptime    float64   `json:"uptime"` // 1 - error_rate; 1.0 when Total == 0
	P95Ms     float64   `json:"p95_ms"` // p95 duration of successful probes; 0 when no successful probes
}

// HistoryReader queries InfluxDB 3 for per-provider time-bucketed history.
type HistoryReader interface {
	ProviderHistory(ctx context.Context, providerID string, window, bucketSize time.Duration) ([]HistoryBucket, error)
}

// HistoryReaderConfig holds connection parameters for InfluxDB 3.
type HistoryReaderConfig struct {
	Host     string
	Token    string
	Database string
}

// NewHistoryReader returns an *influxHistoryReader that satisfies both
// HistoryReader and LiveStatsReader.
func NewHistoryReader(cfg HistoryReaderConfig, client *http.Client) *influxHistoryReader {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	return &influxHistoryReader{cfg: cfg, client: client}
}

type influxHistoryReader struct {
	cfg    HistoryReaderConfig
	client *http.Client
}

func (r *influxHistoryReader) postQuery(ctx context.Context, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		r.cfg.Host+"/api/v3/query_sql", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Token "+r.cfg.Token)
	req.Header.Set("Content-Type", "application/json")
	return r.client.Do(req)
}

func (r *influxHistoryReader) ProviderHistory(
	ctx context.Context,
	providerID string,
	window, bucketSize time.Duration,
) ([]HistoryBucket, error) {
	// Sanitise providerID: only [a-z0-9_-] are valid provider IDs.
	if strings.ContainsAny(providerID, "'\";\\") {
		return nil, fmt.Errorf("influx: invalid provider_id %q", providerID)
	}

	sql := fmt.Sprintf(
		`SELECT date_bin(INTERVAL '%d seconds', time, TIMESTAMP '1970-01-01 00:00:00') AS bucket,
		        COUNT(*) AS total,
		        COUNT(*) FILTER (WHERE success = false) AS errors,
		        COALESCE(approx_percentile_cont(0.95) WITHIN GROUP (ORDER BY duration_ms)
		                 FILTER (WHERE success = true AND duration_ms IS NOT NULL), 0) AS p95_ms
		 FROM probes
		 WHERE time >= now() - INTERVAL '%d seconds'
		   AND provider_id = '%s'
		 GROUP BY bucket
		 ORDER BY bucket`,
		int(bucketSize.Seconds()),
		int(window.Seconds()),
		providerID,
	)

	body, err := json.Marshal(map[string]string{
		"q":      sql,
		"db":     r.cfg.Database,
		"format": "json",
	})
	if err != nil {
		return nil, fmt.Errorf("influx history: marshal query: %w", err)
	}

	resp, err := r.postQuery(ctx, body)
	if err != nil {
		return nil, fmt.Errorf("influx history: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("influx history: status %d: %s", resp.StatusCode, detail)
	}

	var rows []struct {
		Bucket string  `json:"bucket"`
		Total  float64 `json:"total"`
		Errors float64 `json:"errors"`
		P95Ms  float64 `json:"p95_ms"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, fmt.Errorf("influx history: decode: %w", err)
	}

	buckets := make([]HistoryBucket, 0, len(rows))
	for _, row := range rows {
		ts, err := time.Parse(time.RFC3339Nano, row.Bucket)
		if err != nil {
			// Try without nanoseconds.
			ts, err = time.Parse("2006-01-02T15:04:05Z", row.Bucket)
			if err != nil {
				continue
			}
		}
		total := int64(row.Total)
		errors := int64(row.Errors)
		uptime := 1.0
		if total > 0 {
			uptime = 1.0 - float64(errors)/float64(total)
		}
		buckets = append(buckets, HistoryBucket{
			Timestamp: ts.UTC(),
			Total:     total,
			Errors:    errors,
			Uptime:    uptime,
			P95Ms:     row.P95Ms,
		})
	}
	return buckets, nil
}
