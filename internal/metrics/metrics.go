package metrics

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	OpsCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "operations_total",
		Help: "Общее количество операций",
	})

	RateLimitExceededTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "rate_limit_exceeded_total",
		Help: "Количество превышений rate limit",
	})

	GatewayRetriesTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gateway_retries_total",
		Help: "Количество ретраев в gateway",
	})

	HTTPRequestTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Количество HTTP запросов",
		},
		[]string{"method", "path", "status_code"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Длительность HTTP запросов в секундах",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "url", "status_code"},
	)
)

var logger = log.New(os.Stdout, "[INFO] ", log.LstdFlags)

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		statusCode := strconv.Itoa(rw.statusCode)

		routePattern := chi.RouteContext(r.Context()).RoutePattern()
		if routePattern == "" {
			routePattern = "unknown"
		}

		logger.Printf(
			"method=%s path=%s status=%s duration=%dms",
			r.Method,
			routePattern,
			statusCode,
			duration.Milliseconds(),
		)

		HTTPRequestDuration.WithLabelValues(r.Method, routePattern, statusCode).Observe(duration.Seconds())
		HTTPRequestTotal.WithLabelValues(r.Method, routePattern, statusCode).Inc()
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
