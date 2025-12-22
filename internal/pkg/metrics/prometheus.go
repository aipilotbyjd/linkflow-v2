package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTP Metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "linkflow_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "linkflow_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path"},
	)

	// Workflow Execution Metrics
	WorkflowExecutionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "linkflow_workflow_executions_total",
			Help: "Total number of workflow executions",
		},
		[]string{"workspace_id", "status", "trigger_type"},
	)

	WorkflowExecutionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "linkflow_workflow_execution_duration_seconds",
			Help:    "Workflow execution duration in seconds",
			Buckets: []float64{.1, .5, 1, 2.5, 5, 10, 30, 60, 120, 300, 600},
		},
		[]string{"workspace_id", "workflow_id"},
	)

	WorkflowExecutionsInProgress = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "linkflow_workflow_executions_in_progress",
			Help: "Number of workflow executions currently in progress",
		},
		[]string{"workspace_id"},
	)

	// Node Execution Metrics
	NodeExecutionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "linkflow_node_executions_total",
			Help: "Total number of node executions",
		},
		[]string{"node_type", "status"},
	)

	NodeExecutionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "linkflow_node_execution_duration_seconds",
			Help:    "Node execution duration in seconds",
			Buckets: []float64{.01, .05, .1, .25, .5, 1, 2.5, 5, 10, 30},
		},
		[]string{"node_type"},
	)

	// Webhook Metrics
	WebhooksReceivedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "linkflow_webhooks_received_total",
			Help: "Total number of webhooks received",
		},
		[]string{"workspace_id", "path"},
	)

	WebhookQueueDepth = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "linkflow_webhook_queue_depth",
			Help: "Number of webhooks in the queue",
		},
	)

	WebhookDLQDepth = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "linkflow_webhook_dlq_depth",
			Help: "Number of webhooks in dead letter queue",
		},
	)

	// Queue Metrics
	QueueTasksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "linkflow_queue_tasks_total",
			Help: "Total number of tasks enqueued",
		},
		[]string{"task_type"},
	)

	QueueTasksProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "linkflow_queue_tasks_processed_total",
			Help: "Total number of tasks processed",
		},
		[]string{"task_type", "status"},
	)

	QueueDepth = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "linkflow_queue_depth",
			Help: "Number of tasks in the queue",
		},
		[]string{"queue_name"},
	)

	// Database Metrics
	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "linkflow_db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"operation"},
	)

	DBConnectionsOpen = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "linkflow_db_connections_open",
			Help: "Number of open database connections",
		},
	)

	// Rate Limiting Metrics
	RateLimitHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "linkflow_rate_limit_hits_total",
			Help: "Total number of rate limit hits",
		},
		[]string{"workspace_id", "endpoint"},
	)

	// System Metrics
	WorkspacesActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "linkflow_workspaces_active",
			Help: "Number of active workspaces",
		},
	)

	UsersActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "linkflow_users_active",
			Help: "Number of active users (last 24h)",
		},
	)
)

// Handler returns the Prometheus HTTP handler
func Handler() http.Handler {
	return promhttp.Handler()
}

// MetricsMiddleware records HTTP metrics
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start).Seconds()
		path := r.URL.Path

		HTTPRequestsTotal.WithLabelValues(r.Method, path, strconv.Itoa(wrapped.statusCode)).Inc()
		HTTPRequestDuration.WithLabelValues(r.Method, path).Observe(duration)
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

// RecordWorkflowExecution records workflow execution metrics
func RecordWorkflowExecution(workspaceID, workflowID, status, triggerType string, durationSeconds float64) {
	WorkflowExecutionsTotal.WithLabelValues(workspaceID, status, triggerType).Inc()
	if durationSeconds > 0 {
		WorkflowExecutionDuration.WithLabelValues(workspaceID, workflowID).Observe(durationSeconds)
	}
}

// RecordNodeExecution records node execution metrics
func RecordNodeExecution(nodeType, status string, durationSeconds float64) {
	NodeExecutionsTotal.WithLabelValues(nodeType, status).Inc()
	if durationSeconds > 0 {
		NodeExecutionDuration.WithLabelValues(nodeType).Observe(durationSeconds)
	}
}

// RecordWebhook records webhook metrics
func RecordWebhook(workspaceID, path string) {
	WebhooksReceivedTotal.WithLabelValues(workspaceID, path).Inc()
}

// RecordRateLimitHit records rate limit hits
func RecordRateLimitHit(workspaceID, endpoint string) {
	RateLimitHitsTotal.WithLabelValues(workspaceID, endpoint).Inc()
}

// UpdateQueueDepth updates the queue depth gauge
func UpdateQueueDepth(queueName string, depth int64) {
	QueueDepth.WithLabelValues(queueName).Set(float64(depth))
}
