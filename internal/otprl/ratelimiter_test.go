package otprl_test

import (
	"context"
	"testing"

	"github.com/llmstatus/llmstatus/internal/otprl"
)

// Unit tests — MemoryLimiter (no external deps).

func TestMemoryLimiter_AllowsUpToMax(t *testing.T) {
	lim := otprl.NewMemory()
	ctx := context.Background()
	const email = "test@example.com"

	for i := range otprl.MaxAttempts {
		ok, _, err := lim.Allow(ctx, email)
		if err != nil {
			t.Fatalf("attempt %d: unexpected error: %v", i+1, err)
		}
		if !ok {
			t.Fatalf("attempt %d: expected allowed, got denied", i+1)
		}
	}
}

func TestMemoryLimiter_BlocksFourthAttempt(t *testing.T) {
	lim := otprl.NewMemory()
	ctx := context.Background()
	const email = "flood@example.com"

	for range otprl.MaxAttempts {
		lim.Allow(ctx, email) //nolint:errcheck
	}

	ok, retryAfter, err := lim.Allow(ctx, email)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("4th attempt should be denied")
	}
	if retryAfter <= 0 {
		t.Errorf("retryAfter should be positive, got %v", retryAfter)
	}
}

func TestMemoryLimiter_IndependentPerEmail(t *testing.T) {
	lim := otprl.NewMemory()
	ctx := context.Background()

	for range otprl.MaxAttempts {
		lim.Allow(ctx, "a@example.com") //nolint:errcheck
	}
	// b@example.com is a fresh counter
	ok, _, err := lim.Allow(ctx, "b@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("different email should not be rate-limited")
	}
}
