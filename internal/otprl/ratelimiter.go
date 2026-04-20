// Package otprl implements per-email OTP send rate limiting.
package otprl

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	MaxAttempts = 3
	Window      = 10 * time.Minute
)

// Limiter checks whether an OTP send is allowed for a given email.
type Limiter interface {
	// Allow returns (true, zero) when the request is permitted.
	// Returns (false, retryAfter) when the rate limit is exceeded.
	Allow(ctx context.Context, email string) (ok bool, retryAfter time.Duration, err error)
}

// RedisLimiter is a Redis-backed implementation of Limiter.
type RedisLimiter struct {
	rdb *redis.Client
}

func NewRedis(rdb *redis.Client) *RedisLimiter {
	return &RedisLimiter{rdb: rdb}
}

func (r *RedisLimiter) Allow(ctx context.Context, email string) (bool, time.Duration, error) {
	key := otpKey(email)

	// SetNX initialises the counter at 0 with TTL on the very first request.
	// INCR is then always safe: if the key was just created we increment to 1;
	// if it already existed we increment the existing counter.
	r.rdb.SetNX(ctx, key, 0, Window)

	count, err := r.rdb.Incr(ctx, key).Result()
	if err != nil {
		return false, 0, fmt.Errorf("otprl: redis incr: %w", err)
	}
	if count > MaxAttempts {
		ttl, err := r.rdb.TTL(ctx, key).Result()
		if err != nil || ttl < 0 {
			ttl = Window
		}
		return false, ttl, nil
	}
	return true, 0, nil
}

func otpKey(email string) string {
	h := sha256.Sum256([]byte(email))
	return fmt.Sprintf("otp_rate:%x", h)
}

// MemoryLimiter is an in-process Limiter for unit tests.
type MemoryLimiter struct {
	counts map[string]int
}

func NewMemory() *MemoryLimiter {
	return &MemoryLimiter{counts: make(map[string]int)}
}

func (m *MemoryLimiter) Allow(_ context.Context, email string) (bool, time.Duration, error) {
	m.counts[email]++
	if m.counts[email] > MaxAttempts {
		return false, Window, nil
	}
	return true, 0, nil
}
