package testutil

import "testing"

// NewRedis starts an isolated Redis instance via testcontainers and
// returns its address. Cleanup is registered via t.Cleanup.
//
// Scaffolded by LLMS-001. Implementation follows in LLMS-002+.
func NewRedis(t *testing.T) string {
	t.Helper()
	t.Skip("testutil.NewRedis: not implemented (LLMS-002)")
	return ""
}
