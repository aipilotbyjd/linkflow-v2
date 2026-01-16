package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all application metrics
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	HTTPActiveRequests   prometheus.Gauge

	// Workflow metrics
	WorkflowExecutions   *prometheus.CounterVec
	WorkflowDuration     *prometheus.HistogramVec
	WorkflowActive       prometheus.Gauge

	// Node metrics
	NodeExecutions       *prometheus.CounterVec
	NodeDuration         *prometheus.HistogramVec
	NodeRetries          *prometheus.CounterVec

	// Queue metrics
	QueueSize            *prometheus.GaugeVec
	QueueLatency         *prometheus.HistogramVec
	QueueProcessed       *prometheus.CounterVec

	// Database metrics
	DBConnections        prometheus.Gauge
	DBQueryDuration      *prometheus.HistogramVec

	// Cache metrics
	CacheHits            prometheus.Counter
	CacheMisses          prometheus.Counter
}

// New creates and registers all metrics
func New(namespace string) *Metrics {
	m := &Metrics{
		// HTTP metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"method", "path"},
		),
		HTTPActiveRequests: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "http_active_requests",
				Help:      "Number of active HTTP requests",
			},
		),

		// Workflow metrics
		WorkflowExecutions: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "workflow_executions_total",
				Help:      "Total number of workflow executions",
			},
			[]string{"workflow_id", "status", "trigger_type"},
		),
		WorkflowDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "workflow_duration_seconds",
				Help:      "Workflow execution duration in seconds",
				Buckets:   []float64{.1, .25, .5, 1, 2.5, 5, 10, 30, 60, 120, 300},
			},
			[]string{"workflow_id"},
		),
		WorkflowActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "workflow_active_executions",
				Help:      "Number of active workflow executions",
			},
		),

		// Node metrics
		NodeExecutions: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "node_executions_total",
				Help:      "Total number of node executions",
			},
			[]string{"node_type", "status"},
		),
		NodeDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "node_duration_seconds",
				Help:      "Node execution duration in seconds",
				Buckets:   []float64{.01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"node_type"},
		),
		NodeRetries: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "node_retries_total",
				Help:      "Total number of node retries",
			},
			[]string{"node_type"},
		),

		// Queue metrics
		QueueSize: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "queue_size",
				Help:      "Current queue size",
			},
			[]string{"queue"},
		),
		QueueLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "queue_latency_seconds",
				Help:      "Queue processing latency in seconds",
				Buckets:   []float64{.01, .025, .05, .1, .25, .5, 1, 2.5, 5},
			},
			[]string{"queue"},
		),
		QueueProcessed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "queue_processed_total",
				Help:      "Total number of queue items processed",
			},
			[]string{"queue", "status"},
		),

		// Database metrics
		DBConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "db_connections",
				Help:      "Number of database connections",
			},
		),
		DBQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "db_query_duration_seconds",
				Help:      "Database query duration in seconds",
				Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
			},
			[]string{"operation"},
		),

		// Cache metrics
		CacheHits: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cache_hits_total",
				Help:      "Total number of cache hits",
			},
		),
		CacheMisses: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cache_misses_total",
				Help:      "Total number of cache misses",
			},
		),
	}

	return m
}

// RecordHTTPRequest records an HTTP request metric
func (m *Metrics) RecordHTTPRequest(method, path string, status int, duration time.Duration) {
	statusStr := statusLabel(status)
	m.HTTPRequestsTotal.WithLabelValues(method, path, statusStr).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

// RecordWorkflowExecution records a workflow execution metric
func (m *Metrics) RecordWorkflowExecution(workflowID, status, triggerType string, duration time.Duration) {
	m.WorkflowExecutions.WithLabelValues(workflowID, status, triggerType).Inc()
	m.WorkflowDuration.WithLabelValues(workflowID).Observe(duration.Seconds())
}

// RecordNodeExecution records a node execution metric
func (m *Metrics) RecordNodeExecution(nodeType, status string, duration time.Duration) {
	m.NodeExecutions.WithLabelValues(nodeType, status).Inc()
	m.NodeDuration.WithLabelValues(nodeType).Observe(duration.Seconds())
}

// RecordDBQuery records a database query metric
func (m *Metrics) RecordDBQuery(operation string, duration time.Duration) {
	m.DBQueryDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

func statusLabel(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "2xx"
	case code >= 300 && code < 400:
		return "3xx"
	case code >= 400 && code < 500:
		return "4xx"
	case code >= 500:
		return "5xx"
	default:
		return "unknown"
	}
}
