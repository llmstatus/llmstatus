// Command export_region_data exports probe samples for a set of regions from
// InfluxDB 3 as line protocol files, preserving the exact schema the prober
// writes (tags: provider_id/model/probe_type/region_id/error_class; fields:
// success/duration_ms/http_status/tokens_in/tokens_out; ns timestamps).
//
// One-off ops tool for the 2026-08-28 probe-region consolidation. Run locally
// against an SSH tunnel to the main server's InfluxDB (port 18086):
//
//	INFLUX_URL=http://localhost:18086 KEEP_REGIONS=us-west-2,ap-southeast-1 \
//	  go run ./scripts/export_region_data.go
//
// Writes probes-<region>.lp into the current directory (or OUT_DIR).
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	baseURL = getenv("INFLUX_URL", "http://localhost:18086")
	db      = getenv("INFLUX_DB", "probes")
	keep    = strings.Split(getenv("KEEP_REGIONS", "us-west-2,ap-southeast-1"), ",")
	outDir  = getenv("OUT_DIR", ".")
	client  = &http.Client{Timeout: 60 * time.Second}
)

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func main() {
	start := time.Now()
	// Data predates the 2026-08-28 consolidation by months; start conservatively
	// early and let empty days pass through as zero rows (per-day windows are
	// well under InfluxDB 3's parquet file limit, unlike whole-history scans).
	startDate := getenv("START_DATE", "2026-04-01")
	earliest, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		fmt.Fprintln(os.Stderr, "bad START_DATE:", err)
		os.Exit(1)
	}
	fmt.Printf("exporting from %s (kept: %v)\n", earliest.Format(time.RFC3339), keep)

	for _, region := range keep {
		// Append mode: a killed run can resume from START_DATE (first missing
		// day) without losing already-exported days.
		f, err := os.OpenFile(outDir+"/probes-"+region+".lp", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		w := bufio.NewWriterSize(f, 1<<20)

		total := 0
		day := earliest.UTC().Truncate(24 * time.Hour)
		now := time.Now().UTC()
		for day.Before(now) {
			next := day.Add(24 * time.Hour)
			rows, err := queryDay(region, day, next)
			if err != nil {
				fmt.Fprintf(os.Stderr, "query %s %s: %v\n", region, day.Format("2006-01-02"), err)
				os.Exit(1)
			}
			for _, r := range rows {
				line, err := toLineProtocol(r)
				if err != nil {
					fmt.Fprintf(os.Stderr, "convert %s %s: %v\n", region, day.Format("2006-01-02"), err)
					os.Exit(1)
				}
				if _, err := w.WriteString(line + "\n"); err != nil {
					fmt.Fprintln(os.Stderr, "write:", err)
					os.Exit(1)
				}
			}
			total += len(rows)
			if len(rows) > 0 {
				fmt.Printf("  %s %s: %d rows\n", region, day.Format("2006-01-02"), len(rows))
			}
			// Flush after every day so an interrupted run loses at most one day
			// and can be resumed from START_DATE.
			if err := w.Flush(); err != nil {
				fmt.Fprintln(os.Stderr, "flush:", err)
				os.Exit(1)
			}
			day = next
		}
		if err := w.Flush(); err != nil {
			fmt.Fprintln(os.Stderr, "flush:", err)
			os.Exit(1)
		}
		_ = f.Close()
		fmt.Printf("%s: %d rows → probes-%s.lp\n", region, total, region)
	}
	fmt.Printf("done in %s\n", time.Since(start))
}

type probeRow struct {
	ProviderID string `json:"provider_id"`
	Model      string `json:"model"`
	ProbeType  string `json:"probe_type"`
	RegionID   string `json:"region_id"`
	Success    bool   `json:"success"`
	DurationMs int64  `json:"duration_ms"`
	HTTPStatus int64  `json:"http_status"`
	TokensIn   int64  `json:"tokens_in"`
	TokensOut  int64  `json:"tokens_out"`
	ErrorClass string `json:"error_class"`
	Time       string `json:"time"`
}

func queryJSON(q string) ([]probeRow, error) {
	body, err := json.Marshal(map[string]string{"q": q, "db": db, "format": "json"})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, baseURL+"/api/v3/query_sql", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	var rows []probeRow
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("decode: %w (body: %.200s)", err, raw)
	}
	return rows, nil
}

func queryDay(region string, from, to time.Time) ([]probeRow, error) {
	q := fmt.Sprintf(
		"SELECT provider_id, model, probe_type, region_id, success, duration_ms, http_status, tokens_in, tokens_out, error_class, time FROM probes WHERE region_id = '%s' AND time >= '%s' AND time < '%s'",
		region, from.Format(time.RFC3339), to.Format(time.RFC3339))
	return queryJSON(q)
}

func parseTime(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339Nano, s+"Z")
}

// toLineProtocol mirrors internal/store/influx/writer.go exactly.
func toLineProtocol(r probeRow) (string, error) {
	tags := "provider_id=" + escapeTag(r.ProviderID) +
		",model=" + escapeTag(r.Model) +
		",probe_type=" + escapeTag(r.ProbeType) +
		",region_id=" + escapeTag(r.RegionID)
	if r.ErrorClass != "" {
		tags += ",error_class=" + escapeTag(r.ErrorClass)
	}
	success := "false"
	if r.Success {
		success = "true"
	}
	fields := fmt.Sprintf("success=%s,duration_ms=%di,http_status=%di", success, r.DurationMs, r.HTTPStatus)
	if r.TokensIn > 0 {
		fields += fmt.Sprintf(",tokens_in=%di", r.TokensIn)
	}
	if r.TokensOut > 0 {
		fields += fmt.Sprintf(",tokens_out=%di", r.TokensOut)
	}
	ts, err := parseTime(r.Time)
	if err != nil {
		return "", fmt.Errorf("bad time %q: %w", r.Time, err)
	}
	return fmt.Sprintf("probes,%s %s %d", tags, fields, ts.UnixNano()), nil
}

func escapeTag(s string) string {
	s = strings.ReplaceAll(s, ",", `\,`)
	s = strings.ReplaceAll(s, " ", `\ `)
	s = strings.ReplaceAll(s, "=", `\=`)
	return s
}
