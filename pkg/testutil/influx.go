package testutil

import "testing"

// NewInflux starts an isolated InfluxDB 3 instance via testcontainers and
// returns an initialized client handle targeting a clean bucket.
//
// Scaffolded by LLMS-001. Implementation follows in LLMS-002+.
func NewInflux(t *testing.T) string {
	t.Helper()
	t.Skip("testutil.NewInflux: not implemented (LLMS-002)")
	return ""
}
