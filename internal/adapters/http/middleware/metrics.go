package middleware

import (
	"net/http"
	"strconv"
	"time"
)

type MetricsCollector interface {
	RecordHTTPRequest(method, path string, statusCode int, duration time.Duration)
	IncrementActiveRequests()
	DecrementActiveRequests()
}

type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *metricsResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func Metrics(collector MetricsCollector) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			collector.IncrementActiveRequests()
			defer collector.DecrementActiveRequests()

			wrapped := &metricsResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)
			collector.RecordHTTPRequest(r.Method, r.URL.Path, wrapped.statusCode, duration)
		})
	}
}

// NoopMetricsCollector is a no-op implementation
type NoopMetricsCollector struct{}

func (n *NoopMetricsCollector) RecordHTTPRequest(method, path string, statusCode int, duration time.Duration) {
}
func (n *NoopMetricsCollector) IncrementActiveRequests() {}
func (n *NoopMetricsCollector) DecrementActiveRequests() {}

// PrometheusMiddleware wraps prometheus metrics
func PrometheusMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			wrapped := &metricsResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)

			// Set headers for debugging
			w.Header().Set("X-Response-Time", strconv.FormatInt(duration.Milliseconds(), 10)+"ms")
		})
	}
}
