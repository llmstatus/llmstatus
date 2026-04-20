package testutil

import (
	"context"
	"testing"

	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

// NewRedis starts an isolated Redis 7 container and returns its connection URL.
// The container is stopped automatically via t.Cleanup.
func NewRedis(t *testing.T) string {
	t.Helper()
	ctx := context.Background()
	c, err := tcredis.Run(ctx, "redis:7-alpine")
	if err != nil {
		t.Fatalf("testutil.NewRedis: start container: %v", err)
	}
	t.Cleanup(func() {
		if err := c.Terminate(ctx); err != nil {
			t.Logf("testutil.NewRedis: terminate container: %v", err)
		}
	})
	url, err := c.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("testutil.NewRedis: connection string: %v", err)
	}
	return url
}
