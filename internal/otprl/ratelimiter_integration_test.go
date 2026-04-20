//go:build integration

package otprl_test

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"

	"github.com/llmstatus/llmstatus/internal/otprl"
	"github.com/llmstatus/llmstatus/pkg/testutil"
)

func TestRedisLimiter_AllowsUpToMax(t *testing.T) {
	url := testutil.NewRedis(t)
	opt, err := redis.ParseURL(url)
	if err != nil {
		t.Fatalf("parse redis url: %v", err)
	}
	rdb := redis.NewClient(opt)
	t.Cleanup(func() { rdb.Close() })

	lim := otprl.NewRedis(rdb)
	ctx := context.Background()
	const email = "integ@example.com"

	for i := range otprl.MaxAttempts {
		ok, _, err := lim.Allow(ctx, email)
		if err != nil {
			t.Fatalf("attempt %d: %v", i+1, err)
		}
		if !ok {
			t.Fatalf("attempt %d: expected allowed", i+1)
		}
	}
}

func TestRedisLimiter_BlocksFourthAttempt(t *testing.T) {
	url := testutil.NewRedis(t)
	opt, _ := redis.ParseURL(url)
	rdb := redis.NewClient(opt)
	t.Cleanup(func() { rdb.Close() })

	lim := otprl.NewRedis(rdb)
	ctx := context.Background()
	const email = "flood-integ@example.com"

	for range otprl.MaxAttempts {
		lim.Allow(ctx, email) //nolint:errcheck
	}

	ok, retryAfter, err := lim.Allow(ctx, email)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("4th attempt should be denied")
	}
	if retryAfter <= 0 {
		t.Errorf("retryAfter = %v, want > 0", retryAfter)
	}
}
