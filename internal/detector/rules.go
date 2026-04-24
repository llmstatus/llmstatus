package detector

// Detection rule identifiers.
const (
	RuleProviderDown       = "provider_down"
	RuleElevatedErrors     = "elevated_errors"
	RuleLatencyDegradation = "latency_degradation"
	RuleRegionalOutage     = "regional_outage"

	downThreshold     = 0.50
	elevatedThreshold = 0.05
	minProbeCount     = int64(3)

	// Rule 6.3: fire when p95 current > latencyDegradationFactor × p95 baseline.
	latencyDegradationFactor = 3.0
	latencyMinSamples        = int64(5)

	// Rule 6.4: fire when one region exceeds this error rate while not globally down.
	regionalOutageThreshold = 0.50
	minRegionalProbeCount   = int64(3)
)

// Detection is a rule match for a single provider.
type Detection struct {
	ProviderID  string
	Rule        string
	Severity    string
	ErrorRate   float64
	TotalProbes int64
	// Rule 6.3 fields (zero for other rules)
	P95Ms         float64
	BaselineP95Ms float64
	// Rule 6.4 field (empty for other rules)
	Region string
}

// LatencyStats holds p95 latency for one provider over a time window.
// Only successful probes contribute (per METHODOLOGY.md §5.3).
type LatencyStats struct {
	ProviderID  string
	P95Ms       float64
	SampleCount int64
}

// RegionalStats holds error counts for one provider+region over a time window.
type RegionalStats struct {
	ProviderID string
	Region     string
	Total      int64
	Errors     int64
}

// ErrorRate returns the fraction of failed probes, or 0 when Total is 0.
func (s RegionalStats) ErrorRate() float64 {
	if s.Total == 0 {
		return 0
	}
	return float64(s.Errors) / float64(s.Total)
}

// EvaluateRules applies rules 6.1 and 6.2 to the supplied stats.
//
// stats5m  — probe stats over the last 5 minutes  (rule 6.1 window)
// stats10m — probe stats over the last 10 minutes (rule 6.2 window)
//
// Rule 6.1 (DOWN) takes precedence: a provider that is "down" will not also
// appear as "elevated_errors".
func EvaluateRules(stats5m, stats10m []ProbeStats) []Detection {
	// Index 5-minute stats by provider for O(1) lookup.
	byProvider5m := make(map[string]ProbeStats, len(stats5m))
	for _, s := range stats5m {
		byProvider5m[s.ProviderID] = s
	}

	var detections []Detection
	downProviders := make(map[string]struct{})

	// Rule 6.1 — provider is DOWN.
	for _, s := range stats5m {
		if s.Total < minProbeCount {
			continue
		}
		if s.ErrorRate() > downThreshold {
			detections = append(detections, Detection{
				ProviderID:  s.ProviderID,
				Rule:        RuleProviderDown,
				Severity:    "critical",
				ErrorRate:   s.ErrorRate(),
				TotalProbes: s.Total,
			})
			downProviders[s.ProviderID] = struct{}{}
		}
	}

	// Rule 6.2 — elevated errors (skip providers already flagged as down).
	for _, s := range stats10m {
		if _, isDown := downProviders[s.ProviderID]; isDown {
			continue
		}
		if s.Total < minProbeCount {
			continue
		}
		if s.ErrorRate() > elevatedThreshold {
			detections = append(detections, Detection{
				ProviderID:  s.ProviderID,
				Rule:        RuleElevatedErrors,
				Severity:    "major",
				ErrorRate:   s.ErrorRate(),
				TotalProbes: s.Total,
			})
		}
	}

	return detections
}

// EvaluateLatencyRule applies rule 6.3 — degraded latency.
//
// Fires when a provider's p95 latency over `current` exceeds
// latencyDegradationFactor × the same metric over `baseline`.
// Both windows must have at least latencyMinSamples successful probes.
//
// V1 note: baseline is the last 24 h (not the same-hour 7-day median specified
// in METHODOLOGY.md §6). This simplification is tracked in REVIEW_QUEUE.md.
func EvaluateLatencyRule(current, baseline []LatencyStats) []Detection {
	baselineByProvider := make(map[string]LatencyStats, len(baseline))
	for _, b := range baseline {
		baselineByProvider[b.ProviderID] = b
	}

	var detections []Detection
	for _, c := range current {
		if c.SampleCount < latencyMinSamples {
			continue
		}
		b, ok := baselineByProvider[c.ProviderID]
		if !ok || b.SampleCount < latencyMinSamples || b.P95Ms == 0 {
			continue
		}
		if c.P95Ms > latencyDegradationFactor*b.P95Ms {
			detections = append(detections, Detection{
				ProviderID:    c.ProviderID,
				Rule:          RuleLatencyDegradation,
				Severity:      "minor",
				TotalProbes:   c.SampleCount,
				P95Ms:         c.P95Ms,
				BaselineP95Ms: b.P95Ms,
			})
		}
	}
	return detections
}

// EvaluateRegionalRule applies rule 6.4 — regional outage.
//
// Fires when one region of a provider has error rate > regionalOutageThreshold
// while the provider is not already flagged as globally down (rule 6.1).
// Requires at least minRegionalProbeCount probes in the region.
func EvaluateRegionalRule(regional []RegionalStats, globalStats5m []ProbeStats) []Detection {
	downProviders := make(map[string]struct{})
	for _, s := range globalStats5m {
		if s.Total >= minProbeCount && s.ErrorRate() > downThreshold {
			downProviders[s.ProviderID] = struct{}{}
		}
	}

	var detections []Detection
	for _, r := range regional {
		if _, isDown := downProviders[r.ProviderID]; isDown {
			continue
		}
		if r.Total < minRegionalProbeCount {
			continue
		}
		if r.ErrorRate() > regionalOutageThreshold {
			detections = append(detections, Detection{
				ProviderID:  r.ProviderID,
				Rule:        RuleRegionalOutage,
				Severity:    "minor",
				ErrorRate:   r.ErrorRate(),
				TotalProbes: r.Total,
				Region:      r.Region,
			})
		}
	}
	return detections
}
