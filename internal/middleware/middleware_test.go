package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"service-courier/internal/metrics"
	"service-courier/internal/middleware"
	"service-courier/internal/pkg/limiter"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestRateLimitMiddleware_Exceeded(t *testing.T) {
	tb := limiter.NewTokenBucket(0, 0)
	handler := middleware.RateLimitMiddleware(tb)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("handler should not be called when rate limited")
	}))

	before := testutil.ToFloat64(metrics.RateLimitExceededTotal)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d", http.StatusTooManyRequests, rr.Code)
	}

	after := testutil.ToFloat64(metrics.RateLimitExceededTotal)
	if after != before+1 {
		t.Fatalf("expected rate limit metric to increment by 1, before: %v, after: %v", before, after)
	}
}

func TestRateLimitMiddleware_Allowed(t *testing.T) {
	tb := limiter.NewTokenBucket(1, 0)
	handler := middleware.RateLimitMiddleware(tb)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	before := testutil.ToFloat64(metrics.RateLimitExceededTotal)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	after := testutil.ToFloat64(metrics.RateLimitExceededTotal)
	if after != before {
		t.Fatalf("expected rate limit metric to stay the same, before: %v, after: %v", before, after)
	}
}
