// Internal tests for unexported functions in history.go.
package api

import (
	"testing"
	"time"
)

func TestParseWindowParam(t *testing.T) {
	cases := []struct {
		input          string
		wantWindow     time.Duration
		wantBucketSize time.Duration
	}{
		{"24h", 24 * time.Hour, time.Hour},
		{"7d", 7 * 24 * time.Hour, 24 * time.Hour},
		{"30d", 30 * 24 * time.Hour, 24 * time.Hour},
		{"", 30 * 24 * time.Hour, 24 * time.Hour},
		{"bad", 30 * 24 * time.Hour, 24 * time.Hour},
	}
	for _, tc := range cases {
		w, b := parseWindowParam(tc.input)
		if w != tc.wantWindow || b != tc.wantBucketSize {
			t.Errorf("parseWindowParam(%q): got (%v,%v), want (%v,%v)",
				tc.input, w, b, tc.wantWindow, tc.wantBucketSize)
		}
	}
}
