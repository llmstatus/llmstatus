package detector

import (
	"context"
	"testing"
	"time"
)

func TestEvaluateRules_NoProbes(t *testing.T) {
	got := EvaluateRules(nil, nil)
	if len(got) != 0 {
		t.Errorf("expected 0 detections on empty stats, got %d", len(got))
	}
}

func TestEvaluateRules_Down(t *testing.T) {
	stats5m := []ProbeStats{
		{ProviderID: "openai", Total: 10, Errors: 6}, // 60% error rate
	}
	detections := EvaluateRules(stats5m, nil)

	if len(detections) != 1 {
		t.Fatalf("expected 1 detection, got %d", len(detections))
	}
	d := detections[0]
	if d.ProviderID != "openai" {
		t.Errorf("ProviderID: got %q, want openai", d.ProviderID)
	}
	if d.Rule != RuleProviderDown {
		t.Errorf("Rule: got %q, want %q", d.Rule, RuleProviderDown)
	}
	if d.Severity != "critical" {
		t.Errorf("Severity: got %q, want critical", d.Severity)
	}
}

func TestEvaluateRules_Down_TooFewProbes(t *testing.T) {
	// Below minProbeCount — should not fire even with 100% error rate.
	stats5m := []ProbeStats{
		{ProviderID: "openai", Total: 2, Errors: 2},
	}
	got := EvaluateRules(stats5m, nil)
	if len(got) != 0 {
		t.Errorf("expected 0 detections with only 2 probes, got %d", len(got))
	}
}

func TestEvaluateRules_ElevatedErrors(t *testing.T) {
	stats10m := []ProbeStats{
		{ProviderID: "anthropic", Total: 20, Errors: 2}, // 10% error rate
	}
	detections := EvaluateRules(nil, stats10m)

	if len(detections) != 1 {
		t.Fatalf("expected 1 detection, got %d", len(detections))
	}
	d := detections[0]
	if d.Rule != RuleElevatedErrors {
		t.Errorf("Rule: got %q, want %q", d.Rule, RuleElevatedErrors)
	}
	if d.Severity != "major" {
		t.Errorf("Severity: got %q, want major", d.Severity)
	}
}

func TestEvaluateRules_Down_SuppressesElevated(t *testing.T) {
	// Provider appears in both 5m and 10m stats — down rule wins.
	stats5m := []ProbeStats{
		{ProviderID: "openai", Total: 10, Errors: 6}, // 60% — DOWN
	}
	stats10m := []ProbeStats{
		{ProviderID: "openai", Total: 20, Errors: 4}, // 20% — would be elevated
	}
	detections := EvaluateRules(stats5m, stats10m)

	if len(detections) != 1 {
		t.Fatalf("expected exactly 1 detection (down, not both), got %d", len(detections))
	}
	if detections[0].Rule != RuleProviderDown {
		t.Errorf("expected provider_down rule, got %q", detections[0].Rule)
	}
}

func TestEvaluateRules_Operational(t *testing.T) {
	stats5m := []ProbeStats{
		{ProviderID: "openai", Total: 10, Errors: 0},
	}
	stats10m := []ProbeStats{
		{ProviderID: "openai", Total: 20, Errors: 1}, // 5% — at threshold, not above
	}
	got := EvaluateRules(stats5m, stats10m)
	if len(got) != 0 {
		t.Errorf("expected 0 detections at 5%% error rate, got %d", len(got))
	}
}

func TestEvaluateRules_MultipleProviders(t *testing.T) {
	stats5m := []ProbeStats{
		{ProviderID: "openai", Total: 10, Errors: 6},    // DOWN
		{ProviderID: "anthropic", Total: 10, Errors: 0}, // OK
	}
	stats10m := []ProbeStats{
		{ProviderID: "deepseek", Total: 10, Errors: 2}, // elevated (20%)
	}
	detections := EvaluateRules(stats5m, stats10m)

	if len(detections) != 2 {
		t.Fatalf("expected 2 detections (openai down + deepseek elevated), got %d", len(detections))
	}
}

// ---- Rule 6.3 tests -----------------------------------------------------------

func TestEvaluateLatencyRule_Fires(t *testing.T) {
	current := []LatencyStats{{ProviderID: "openai", P95Ms: 6000, SampleCount: 10}}
	baseline := []LatencyStats{{ProviderID: "openai", P95Ms: 1000, SampleCount: 20}}
	// 6000 > 3 × 1000 → should fire
	detections := EvaluateLatencyRule(current, baseline)
	if len(detections) != 1 {
		t.Fatalf("expected 1 detection, got %d", len(detections))
	}
	d := detections[0]
	if d.Rule != RuleLatencyDegradation {
		t.Errorf("Rule: got %q, want %q", d.Rule, RuleLatencyDegradation)
	}
	if d.Severity != "minor" {
		t.Errorf("Severity: got %q, want minor", d.Severity)
	}
	if d.P95Ms != 6000 || d.BaselineP95Ms != 1000 {
		t.Errorf("P95Ms=%f BaselineP95Ms=%f", d.P95Ms, d.BaselineP95Ms)
	}
}

func TestEvaluateLatencyRule_BelowThreshold(t *testing.T) {
	// 2500 < 3 × 1000 → should NOT fire
	current := []LatencyStats{{ProviderID: "openai", P95Ms: 2500, SampleCount: 10}}
	baseline := []LatencyStats{{ProviderID: "openai", P95Ms: 1000, SampleCount: 20}}
	got := EvaluateLatencyRule(current, baseline)
	if len(got) != 0 {
		t.Errorf("expected no detections at 2.5× baseline, got %d", len(got))
	}
}

func TestEvaluateLatencyRule_TooFewCurrentSamples(t *testing.T) {
	current := []LatencyStats{{ProviderID: "openai", P95Ms: 9999, SampleCount: 2}} // < 5
	baseline := []LatencyStats{{ProviderID: "openai", P95Ms: 100, SampleCount: 20}}
	got := EvaluateLatencyRule(current, baseline)
	if len(got) != 0 {
		t.Errorf("expected no detection with too few current samples, got %d", len(got))
	}
}

func TestEvaluateLatencyRule_TooFewBaselineSamples(t *testing.T) {
	current := []LatencyStats{{ProviderID: "openai", P95Ms: 9999, SampleCount: 10}}
	baseline := []LatencyStats{{ProviderID: "openai", P95Ms: 100, SampleCount: 1}} // < 5
	got := EvaluateLatencyRule(current, baseline)
	if len(got) != 0 {
		t.Errorf("expected no detection with too few baseline samples, got %d", len(got))
	}
}

func TestEvaluateLatencyRule_NoBaseline(t *testing.T) {
	current := []LatencyStats{{ProviderID: "openai", P95Ms: 9999, SampleCount: 10}}
	got := EvaluateLatencyRule(current, nil)
	if len(got) != 0 {
		t.Errorf("expected no detection when no baseline exists, got %d", len(got))
	}
}

func TestEvaluateLatencyRule_ZeroBaselineP95(t *testing.T) {
	current := []LatencyStats{{ProviderID: "openai", P95Ms: 9999, SampleCount: 10}}
	baseline := []LatencyStats{{ProviderID: "openai", P95Ms: 0, SampleCount: 10}}
	got := EvaluateLatencyRule(current, baseline)
	if len(got) != 0 {
		t.Errorf("expected no detection with zero baseline p95, got %d", len(got))
	}
}

// ---- Rule 6.4 tests -----------------------------------------------------------

func TestEvaluateRegionalRule_Fires(t *testing.T) {
	regional := []RegionalStats{
		{ProviderID: "openai", Region: "us-east-1", Total: 10, Errors: 6},  // 60%
		{ProviderID: "openai", Region: "eu-central", Total: 10, Errors: 0}, // 0% — healthy
	}
	// Global stats show openai is NOT down overall.
	global := []ProbeStats{{ProviderID: "openai", Total: 20, Errors: 6}} // 30%
	detections := EvaluateRegionalRule(regional, global)
	if len(detections) != 1 {
		t.Fatalf("expected 1 detection, got %d: %+v", len(detections), detections)
	}
	d := detections[0]
	if d.Rule != RuleRegionalOutage {
		t.Errorf("Rule: got %q, want %q", d.Rule, RuleRegionalOutage)
	}
	if d.Severity != "minor" {
		t.Errorf("Severity: got %q, want minor", d.Severity)
	}
	if d.Region != "us-east-1" {
		t.Errorf("Region: got %q, want us-east-1", d.Region)
	}
}

func TestEvaluateRegionalRule_GloballyDown_Suppressed(t *testing.T) {
	// Provider is already globally DOWN — regional rule must not fire.
	regional := []RegionalStats{
		{ProviderID: "openai", Region: "us-east-1", Total: 10, Errors: 9},
	}
	global := []ProbeStats{{ProviderID: "openai", Total: 10, Errors: 6}} // 60% → down
	got := EvaluateRegionalRule(regional, global)
	if len(got) != 0 {
		t.Errorf("expected no detection when provider is globally down, got %d", len(got))
	}
}

func TestEvaluateRegionalRule_BelowThreshold(t *testing.T) {
	regional := []RegionalStats{
		{ProviderID: "openai", Region: "us-east-1", Total: 10, Errors: 4}, // 40% < 50%
	}
	got := EvaluateRegionalRule(regional, nil)
	if len(got) != 0 {
		t.Errorf("expected no detection below threshold, got %d", len(got))
	}
}

func TestEvaluateRegionalRule_TooFewProbes(t *testing.T) {
	regional := []RegionalStats{
		{ProviderID: "openai", Region: "us-east-1", Total: 2, Errors: 2}, // < 3
	}
	got := EvaluateRegionalRule(regional, nil)
	if len(got) != 0 {
		t.Errorf("expected no detection with too few regional probes, got %d", len(got))
	}
}

func TestEvaluateRegionalRule_MultipleRegions(t *testing.T) {
	regional := []RegionalStats{
		{ProviderID: "openai", Region: "us-east-1", Total: 10, Errors: 6},      // 60% — fire
		{ProviderID: "openai", Region: "ap-northeast-1", Total: 10, Errors: 9}, // 90% — fire
		{ProviderID: "openai", Region: "eu-central", Total: 10, Errors: 0},     // healthy
	}
	global := []ProbeStats{{ProviderID: "openai", Total: 30, Errors: 15}} // 50% — exactly at threshold, NOT > threshold
	detections := EvaluateRegionalRule(regional, global)
	if len(detections) != 2 {
		t.Fatalf("expected 2 regional detections, got %d", len(detections))
	}
}

func TestRegionalStats_ErrorRate(t *testing.T) {
	cases := []struct {
		total, errors int64
		want          float64
	}{
		{0, 0, 0},
		{10, 5, 0.5},
		{10, 0, 0},
	}
	for _, tc := range cases {
		s := RegionalStats{Total: tc.total, Errors: tc.errors}
		if got := s.ErrorRate(); got != tc.want {
			t.Errorf("ErrorRate(%d/%d) = %f, want %f", tc.errors, tc.total, got, tc.want)
		}
	}
}

// ---- Rule 6.3 + 6.4 runner integration ----------------------------------------

func TestRunner_LatencyDegradation_CreatesIncident(t *testing.T) {
	store := &fakeIncidentStore{}
	r := New(&latencyFakeReader{
		current:  []LatencyStats{{ProviderID: "openai", P95Ms: 9000, SampleCount: 10}},
		baseline: []LatencyStats{{ProviderID: "openai", P95Ms: 1000, SampleCount: 20}},
	}, store, time.Hour)
	r.runOnce(context.Background())

	found := false
	for _, inc := range store.created {
		if inc.DetectionRule.String == RuleLatencyDegradation {
			found = true
			if inc.Severity != "minor" {
				t.Errorf("Severity: got %q, want minor", inc.Severity)
			}
		}
	}
	if !found {
		t.Fatal("expected latency_degradation incident to be created")
	}
}

func TestRunner_RegionalOutage_CreatesIncident(t *testing.T) {
	store := &fakeIncidentStore{}
	r := New(&regionalFakeReader{
		regional: []RegionalStats{
			{ProviderID: "anthropic", Region: "us-east-1", Total: 5, Errors: 4}, // 80%
		},
	}, store, time.Hour)
	r.runOnce(context.Background())

	found := false
	for _, inc := range store.created {
		if inc.DetectionRule.String == RuleRegionalOutage {
			found = true
			if inc.Severity != "minor" {
				t.Errorf("Severity: got %q, want minor", inc.Severity)
			}
		}
	}
	if !found {
		t.Fatal("expected regional_outage incident to be created")
	}
}

// ---- specialized test readers --------------------------------------------------

// latencyFakeReader returns distinct data for current (5m) vs baseline (24h).
type latencyFakeReader struct {
	current  []LatencyStats
	baseline []LatencyStats
}

func (f *latencyFakeReader) ErrorRateByProvider(_ context.Context, _ time.Duration) ([]ProbeStats, error) {
	return nil, nil
}
func (f *latencyFakeReader) LatencyByProvider(_ context.Context, window time.Duration) ([]LatencyStats, error) {
	if window <= 5*time.Minute {
		return f.current, nil
	}
	return f.baseline, nil
}
func (f *latencyFakeReader) RegionalErrorRateByProvider(_ context.Context, _ time.Duration) ([]RegionalStats, error) {
	return nil, nil
}
func (f *latencyFakeReader) QualityByProvider(_ context.Context, _ time.Duration) ([]ProbeStats, error) {
	return nil, nil
}

// regionalFakeReader only populates regional stats.
type regionalFakeReader struct {
	regional []RegionalStats
}

func (f *regionalFakeReader) ErrorRateByProvider(_ context.Context, _ time.Duration) ([]ProbeStats, error) {
	return nil, nil
}
func (f *regionalFakeReader) LatencyByProvider(_ context.Context, _ time.Duration) ([]LatencyStats, error) {
	return nil, nil
}
func (f *regionalFakeReader) RegionalErrorRateByProvider(_ context.Context, _ time.Duration) ([]RegionalStats, error) {
	return f.regional, nil
}
func (f *regionalFakeReader) QualityByProvider(_ context.Context, _ time.Duration) ([]ProbeStats, error) {
	return nil, nil
}

// qualityFakeReader returns distinct data for current (5m) vs baseline (24h).
type qualityFakeReader struct {
	current  []ProbeStats
	baseline []ProbeStats
}

func (f *qualityFakeReader) ErrorRateByProvider(_ context.Context, _ time.Duration) ([]ProbeStats, error) {
	return nil, nil
}
func (f *qualityFakeReader) LatencyByProvider(_ context.Context, _ time.Duration) ([]LatencyStats, error) {
	return nil, nil
}
func (f *qualityFakeReader) RegionalErrorRateByProvider(_ context.Context, _ time.Duration) ([]RegionalStats, error) {
	return nil, nil
}
func (f *qualityFakeReader) QualityByProvider(_ context.Context, window time.Duration) ([]ProbeStats, error) {
	if window <= 5*time.Minute {
		return f.current, nil
	}
	return f.baseline, nil
}

func TestProbeStats_ErrorRate(t *testing.T) {
	cases := []struct {
		total, errors int64
		want          float64
	}{
		{0, 0, 0},
		{10, 5, 0.5},
		{10, 0, 0},
		{10, 10, 1.0},
	}
	for _, tc := range cases {
		s := ProbeStats{Total: tc.total, Errors: tc.errors}
		if got := s.ErrorRate(); got != tc.want {
			t.Errorf("ErrorRate(%d/%d) = %f, want %f", tc.errors, tc.total, got, tc.want)
		}
	}
}

// ---- Rule 6.5 — quality degradation -----------------------------------------

func TestEvaluateQualityRule_NoProbes(t *testing.T) {
	got := EvaluateQualityRule(nil, nil)
	if len(got) != 0 {
		t.Errorf("expected 0 detections on empty stats, got %d", len(got))
	}
}

func TestEvaluateQualityRule_TooFewProbes(t *testing.T) {
	current := []ProbeStats{{ProviderID: "openai", Total: 1, Errors: 1}}
	got := EvaluateQualityRule(current, nil)
	if len(got) != 0 {
		t.Errorf("expected 0 detections with only 1 probe, got %d", len(got))
	}
}

func TestEvaluateQualityRule_AbsoluteThreshold_NoBaseline(t *testing.T) {
	// 3 of 5 quality probes fail (60%) — exceeds 30% absolute threshold with no baseline data.
	current := []ProbeStats{{ProviderID: "openai", Total: 5, Errors: 3}}
	got := EvaluateQualityRule(current, nil)
	if len(got) != 1 {
		t.Fatalf("expected 1 detection, got %d", len(got))
	}
	d := got[0]
	if d.Rule != RuleQualityDegradation {
		t.Errorf("Rule: got %q, want %q", d.Rule, RuleQualityDegradation)
	}
	if d.Severity != "minor" {
		t.Errorf("Severity: got %q, want minor", d.Severity)
	}
}

func TestEvaluateQualityRule_BelowAbsoluteThreshold(t *testing.T) {
	// 1 of 5 quality probes fail (20%) — below 30% absolute threshold.
	current := []ProbeStats{{ProviderID: "openai", Total: 5, Errors: 1}}
	got := EvaluateQualityRule(current, nil)
	if len(got) != 0 {
		t.Errorf("expected no detection below absolute threshold, got %d", len(got))
	}
}

func TestEvaluateQualityRule_RelativeThreshold_Fires(t *testing.T) {
	// Baseline: 1/20 = 5%. Current: 4/5 = 80% → 16× baseline, exceeds 3× threshold.
	current := []ProbeStats{{ProviderID: "anthropic", Total: 5, Errors: 4}}
	baseline := []ProbeStats{{ProviderID: "anthropic", Total: 20, Errors: 1}}
	got := EvaluateQualityRule(current, baseline)
	if len(got) != 1 {
		t.Fatalf("expected 1 detection, got %d", len(got))
	}
	if got[0].Rule != RuleQualityDegradation {
		t.Errorf("Rule: got %q, want %q", got[0].Rule, RuleQualityDegradation)
	}
}

func TestEvaluateQualityRule_RelativeThreshold_BelowFactor(t *testing.T) {
	// Baseline: 2/20 = 10%. Current: 1/5 = 20% → 2× baseline, below 3× threshold.
	current := []ProbeStats{{ProviderID: "anthropic", Total: 5, Errors: 1}}
	baseline := []ProbeStats{{ProviderID: "anthropic", Total: 20, Errors: 2}}
	got := EvaluateQualityRule(current, baseline)
	if len(got) != 0 {
		t.Errorf("expected no detection when ratio below factor, got %d", len(got))
	}
}

func TestRunner_QualityDegradation_CreatesIncident(t *testing.T) {
	store := &fakeIncidentStore{}
	r := New(&qualityFakeReader{
		current:  []ProbeStats{{ProviderID: "openai", Total: 5, Errors: 3}}, // 60% failure
		baseline: []ProbeStats{},                                             // no baseline → absolute threshold applies
	}, store, time.Hour)
	r.runOnce(context.Background())

	found := false
	for _, inc := range store.created {
		if inc.DetectionRule.String == RuleQualityDegradation {
			found = true
			if inc.Severity != "minor" {
				t.Errorf("Severity: got %q, want minor", inc.Severity)
			}
		}
	}
	if !found {
		t.Fatal("expected quality_degradation incident to be created")
	}
}
