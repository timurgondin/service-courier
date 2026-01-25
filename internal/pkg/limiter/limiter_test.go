package limiter_test

import (
	"testing"
	"time"

	"service-courier/internal/pkg/limiter"
)

func TestTokenBucketAllow(t *testing.T) {
	tb := limiter.NewTokenBucket(2, 1)

	if !tb.Allow() {
		t.Fatalf("expected first Allow to succeed")
	}
	if !tb.Allow() {
		t.Fatalf("expected second Allow to succeed")
	}
	if tb.Allow() {
		t.Fatalf("expected third Allow to be rate limited")
	}
}

func TestTokenBucketRefill(t *testing.T) {
	tb := limiter.NewTokenBucket(1, 1)

	if !tb.Allow() {
		t.Fatalf("expected Allow to succeed")
	}
	if tb.Allow() {
		t.Fatalf("expected bucket to be empty")
	}

	time.Sleep(1100 * time.Millisecond)

	if !tb.Allow() {
		t.Fatalf("expected token to be refilled after sleep")
	}
}
