package influx

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// ProviderLiveStat holds current 24 h aggregate stats for one provider.
type ProviderLiveStat struct {
	ProviderID string
	Uptime24h  float64 // 0–1; 1.0 when Total == 0
	P95Ms      float64 // p95 duration of successful probes; 0 when no successful probes
}

// LiveStatsReader fetches current aggregate stats for every active provider
// in a single InfluxDB round-trip.
type LiveStatsReader interface {
	AllProviderLiveStats(ctx context.Context) ([]ProviderLiveStat, error)
}

// compile-time check: *influxHistoryReader satisfies LiveStatsReader.
var _ LiveStatsReader = (*influxHistoryReader)(nil)

// AllProviderLiveStats queries the last 24 h of probes, grouped by provider_id,
// returning uptime and p95 latency for each.
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

	body, err := json.Marshal(map[string]string{
		"q":      sql,
		"db":     r.cfg.Database,
		"format": "json",
	})
	if err != nil {
		return nil, fmt.Errorf("influx live stats: marshal query: %w", err)
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
