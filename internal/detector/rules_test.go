package detector

import (
	"testing"
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
