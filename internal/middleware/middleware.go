package middleware

import (
	"log"
	"net/http"
	tb "service-courier/internal/pkg/limiter"
	"service-courier/internal/metrics"
)

func RateLimitMiddleware(limiter *tb.TokenBucket) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				log.Printf("Rate limit exceeded for %s", r.URL.Path)
				metrics.RateLimitExceededTotal.Inc()
				w.Header().Set("X-RateLimit-Limit", "10")
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte("Rate limit exceeded"))
				return
			}

			log.Printf("Request allowed: %s", r.URL.Path)
			next.ServeHTTP(w, r)
		})
	}
}
