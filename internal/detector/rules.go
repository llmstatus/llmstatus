package detector

const (
	RuleProviderDown    = "provider_down"
	RuleElevatedErrors  = "elevated_errors"

	downThreshold     = 0.50
	elevatedThreshold = 0.05
	minProbeCount     = int64(3) // require at least 3 probes before firing
)

// Detection is a rule match for a single provider.
type Detection struct {
	ProviderID  string
	Rule        string // RuleProviderDown | RuleElevatedErrors
	Severity    string // "critical" | "major"
	ErrorRate   float64
	TotalProbes int64
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
