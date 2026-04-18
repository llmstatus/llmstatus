// Package testutil provides shared helpers for integration tests.
//
// All helpers are expected to be safe to call from t.Parallel tests and
// must register t.Cleanup so no state leaks between tests.
//
// Scaffolded by LLMS-001. Implementations follow in LLMS-002+.
package testutil

import "testing"

// NewPostgres starts an isolated PostgreSQL instance via testcontainers
// and returns a connection string along with a migrated schema.
//
// The caller should treat the returned string as an opaque DSN.
// Cleanup is registered via t.Cleanup.
func NewPostgres(t *testing.T) string {
	t.Helper()
	t.Skip("testutil.NewPostgres: not implemented (LLMS-002)")
	return ""
}
