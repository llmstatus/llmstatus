package testutil

import (
	"net/http/httptest"
	"testing"
)

// NewUpstreamServer starts an httptest.Server used as a mock AI-provider
// upstream. The returned server is closed automatically via t.Cleanup.
//
// Scaffolded by LLMS-001. Implementation follows in LLMS-002+.
func NewUpstreamServer(t *testing.T) *httptest.Server {
	t.Helper()
	t.Skip("testutil.NewUpstreamServer: not implemented (LLMS-002)")
	return nil
}
