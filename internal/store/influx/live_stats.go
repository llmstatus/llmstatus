package influx

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

const (
	sparklineBuckets    = 60
	sparklineBucketSecs = 1440 // 24 min per bucket → 60 × 24 min = 24 h
	windowSecs          = sparklineBuckets * sparklineBucketSecs
)

// ProviderLiveStat holds current 24 h aggregate stats for one provider.
type ProviderLiveStat struct {
	ProviderID string
	Uptime24h  float64 // 0–1; 1.0 when Total == 0
	P95Ms      float64 // p95 duration of successful probes; 0 when no successful probes
}

// ModelLiveStat holds current 24 h aggregate stats for one provider+model pair.
type ModelLiveStat struct {
	ProviderID string
	Model      string
	Uptime24h  float64 // 0–1; 1.0 when Total == 0
	P95Ms      float64 // p95 duration of successful probes; 0 when none
}

// LiveStatsReader fetches current aggregate stats for every active provider
// and model in a single InfluxDB round-trip each.
type LiveStatsReader interface {
	AllProviderLiveStats(ctx context.Context) ([]ProviderLiveStat, error)
	// AllModelLiveStats returns 24 h uptime + p95 grouped by provider_id + model.
	AllModelLiveStats(ctx context.Context) ([]ModelLiveStat, error)
	// AllModelSparklines returns 60-bucket avg latency (ms) per provider+model.
	// Key: "provider_id:model". Value: slice of length 60; 0 means no data.
	AllModelSparklines(ctx context.Context) (map[string][]float64, error)
}

// compile-time check: *influxHistoryReader satisfies LiveStatsReader.
var _ LiveStatsReader = (*influxHistoryReader)(nil)

func (r *influxHistoryReader) AllProviderLiveStats(ctx context.Context) ([]ProviderLiveStat, error) {
	sql := `SELECT
	    provider_id,
	    COUNT(*) AS total,
	    COUNT(*) FILTER (WHERE success = false) AS errors,
	    COALESCE(approx_percentile_cont(0.95) WITHIN GROUP (ORDER BY duration_ms)
	             FILTER (WHERE success = true AND duration_ms IS NOT NULL), 0) AS p95_ms
	FROM probes
	WHERE time >= now() - INTERVAL '86400 seconds'
	GROUP BY provider_id`

	body, err := json.Marshal(map[string]string{"q": sql, "db": r.cfg.Database, "format": "json"})
	if err != nil {
		return nil, fmt.Errorf("influx live stats: marshal: %w", err)
	}

	resp, err := r.postQuery(ctx, body)
	if err != nil {
		return nil, fmt.Errorf("influx live stats: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("influx live stats: status %d: %s", resp.StatusCode, detail)
	}

	var rows []struct {
		ProviderID string  `json:"provider_id"`
		Total      float64 `json:"total"`
		Errors     float64 `json:"errors"`
		P95Ms      float64 `json:"p95_ms"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, fmt.Errorf("influx live stats: decode: %w", err)
	}

	out := make([]ProviderLiveStat, 0, len(rows))
	for _, row := range rows {
		uptime := 1.0
		if row.Total > 0 {
			uptime = 1.0 - row.Errors/row.Total
		}
		out = append(out, ProviderLiveStat{
			ProviderID: row.ProviderID,
			Uptime24h:  uptime,
			P95Ms:      row.P95Ms,
		})
	}
	return out, nil
}

func (r *influxHistoryReader) AllModelLiveStats(ctx context.Context) ([]ModelLiveStat, error) {
	sql := `SELECT
	    provider_id,
	    model,
	    COUNT(*) AS total,
	    COUNT(*) FILTER (WHERE success = false) AS errors,
	    COALESCE(approx_percentile_cont(0.95) WITHIN GROUP (ORDER BY duration_ms)
	             FILTER (WHERE success = true AND duration_ms IS NOT NULL), 0) AS p95_ms
	FROM probes
	WHERE time >= now() - INTERVAL '86400 seconds'
	GROUP BY provider_id, model`

	body, err := json.Marshal(map[string]string{"q": sql, "db": r.cfg.Database, "format": "json"})
	if err != nil {
		return nil, fmt.Errorf("influx model stats: marshal: %w", err)
	}

	resp, err := r.postQuery(ctx, body)
	if err != nil {
		return nil, fmt.Errorf("influx model stats: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("influx model stats: status %d: %s", resp.StatusCode, detail)
	}

	var rows []struct {
		ProviderID string  `json:"provider_id"`
		Model      string  `json:"model"`
		Total      float64 `json:"total"`
		Errors     float64 `json:"errors"`
		P95Ms      float64 `json:"p95_ms"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, fmt.Errorf("influx model stats: decode: %w", err)
	}

	out := make([]ModelLiveStat, 0, len(rows))
	for _, row := range rows {
		uptime := 1.0
		if row.Total > 0 {
			uptime = 1.0 - row.Errors/row.Total
		}
		out = append(out, ModelLiveStat{
			ProviderID: row.ProviderID,
			Model:      row.Model,
			Uptime24h:  uptime,
			P95Ms:      row.P95Ms,
		})
	}
	return out, nil
}

// AllModelSparklines returns 60 avg_ms values per provider+model for the last
// 24 h. Each bucket covers 24 min. Zero means no successful probe data.
// Key format: "provider_id:model".
func (r *influxHistoryReader) AllModelSparklines(ctx context.Context) (map[string][]float64, error) {
	sql := fmt.Sprintf(
		`SELECT
		    provider_id,
		    model,
		    date_bin(INTERVAL '%d seconds', time, TIMESTAMP '1970-01-01 00:00:00') AS bucket,
		    AVG(duration_ms) FILTER (WHERE success = true) AS avg_ms
		 FROM probes
		 WHERE time >= now() - INTERVAL '%d seconds'
		 GROUP BY provider_id, model, bucket
		 ORDER BY provider_id, model, bucket`,
		sparklineBucketSecs,
		windowSecs,
	)

	body, err := json.Marshal(map[string]string{"q": sql, "db": r.cfg.Database, "format": "json"})
	if err != nil {
		return nil, fmt.Errorf("influx sparklines: marshal: %w", err)
	}

	resp, err := r.postQuery(ctx, body)
	if err != nil {
		return nil, fmt.Errorf("influx sparklines: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("influx sparklines: status %d: %s", resp.StatusCode, detail)
	}

	// avg_ms can be null in JSON when the FILTER removes all rows.
	var rows []struct {
		ProviderID string   `json:"provider_id"`
		Model      string   `json:"model"`
		Bucket     string   `json:"bucket"`
		AvgMs      *float64 `json:"avg_ms"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, fmt.Errorf("influx sparklines: decode: %w", err)
	}

	// Compute the first bucket timestamp for the 24 h window.
	nowSec := time.Now().Unix()
	windowStartSec := nowSec - int64(windowSecs)
	firstBucket := (windowStartSec / sparklineBucketSecs) * sparklineBucketSecs

	out := make(map[string][]float64)
	for _, row := range rows {
		ts, err := parseInfluxTime(row.Bucket)
		if err != nil {
			continue
		}
		idx := int((ts.Unix() - firstBucket) / sparklineBucketSecs)
		if idx < 0 || idx >= sparklineBuckets {
			continue
		}
		key := row.ProviderID + ":" + row.Model
		if out[key] == nil {
			out[key] = make([]float64, sparklineBuckets)
		}
		if row.AvgMs != nil {
			out[key][idx] = *row.AvgMs
		}
	}
	return out, nil
}

// parseInfluxTime handles InfluxDB 3's varying timestamp formats:
// RFC3339Nano, with-Z, and without-Z (bare local-time string, treated as UTC).
func parseInfluxTime(s string) (time.Time, error) {
	for _, layout := range []string{
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse influx time %q", s)
}
