package testutil

import "testing"

// FixtureProvider inserts a representative Provider row and returns its
// primary key. Intended for integration tests that need a foreign-key
// parent without bespoke setup.
//
// Scaffolded by LLMS-001. Implementation follows in LLMS-002+.
func FixtureProvider(t *testing.T, dsn string) int64 {
	t.Helper()
	t.Skip("testutil.FixtureProvider: not implemented (LLMS-002)")
	return 0
}

// FixtureChannel inserts a Channel row for the given provider and returns
// its primary key.
//
// Scaffolded by LLMS-001. Implementation follows in LLMS-002+.
func FixtureChannel(t *testing.T, dsn string, providerID int64) int64 {
	t.Helper()
	t.Skip("testutil.FixtureChannel: not implemented (LLMS-002)")
	return 0
}
