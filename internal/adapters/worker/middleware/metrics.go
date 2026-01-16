package middleware

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	taskProcessedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "linkflow_worker_tasks_processed_total",
			Help: "Total number of tasks processed",
		},
		[]string{"task_type", "status"},
	)

	taskDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "linkflow_worker_task_duration_seconds",
			Help:    "Duration of task processing",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
		},
		[]string{"task_type"},
	)

	taskRetries = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "linkflow_worker_task_retries_total",
			Help: "Total number of task retries",
		},
		[]string{"task_type"},
	)
)

// Metrics returns a middleware that collects task metrics
func Metrics() asynq.MiddlewareFunc {
	return func(h asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			start := time.Now()

			retried, _ := asynq.GetRetryCount(ctx)
			if retried > 0 {
				taskRetries.WithLabelValues(t.Type()).Inc()
			}

			err := h.ProcessTask(ctx, t)

			duration := time.Since(start).Seconds()
			taskDuration.WithLabelValues(t.Type()).Observe(duration)

			status := "success"
			if err != nil {
				status = "error"
			}
			taskProcessedTotal.WithLabelValues(t.Type(), status).Inc()

			return err
		})
	}
}
