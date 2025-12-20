package middleware

import (
	"context"
	"sync"
	"time"

	"github.com/linkflow-ai/linkflow/internal/worker/processor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MetricsMiddleware collects Prometheus metrics for node execution
type MetricsMiddleware struct {
	nodeExecutionsTotal *prometheus.CounterVec
	nodeDuration        *prometheus.HistogramVec
	nodeErrorsTotal     *prometheus.CounterVec
	activeNodes         *prometheus.GaugeVec
}

// NewMetricsMiddleware creates a new metrics middleware
func NewMetricsMiddleware() *MetricsMiddleware {
	return &MetricsMiddleware{
		nodeExecutionsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "linkflow_node_executions_total",
				Help: "Total number of node executions",
			},
			[]string{"workspace_id", "node_type", "status"},
		),
		nodeDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "linkflow_node_duration_seconds",
				Help:    "Duration of node executions in seconds",
				Buckets: prometheus.ExponentialBuckets(0.001, 2, 15), // 1ms to ~16s
			},
			[]string{"workspace_id", "node_type"},
		),
		nodeErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "linkflow_node_errors_total",
				Help: "Total number of node execution errors",
			},
			[]string{"workspace_id", "node_type", "error_type"},
		),
		activeNodes: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "linkflow_active_nodes",
				Help: "Number of currently executing nodes",
			},
			[]string{"workspace_id", "node_type"},
		),
	}
}

// Execute implements Middleware
func (m *MetricsMiddleware) Execute(ctx context.Context, rctx *processor.RuntimeContext, node *processor.NodeDefinition, next NextFunc) (map[string]interface{}, error) {
	workspaceID := rctx.WorkspaceID.String()
	nodeType := node.Type

	// Track active nodes
	m.activeNodes.With(prometheus.Labels{
		"workspace_id": workspaceID,
		"node_type":    nodeType,
	}).Inc()
	defer m.activeNodes.With(prometheus.Labels{
		"workspace_id": workspaceID,
		"node_type":    nodeType,
	}).Dec()

	// Track duration
	startTime := time.Now()
	result, err := next(ctx)
	duration := time.Since(startTime)

	m.nodeDuration.With(prometheus.Labels{
		"workspace_id": workspaceID,
		"node_type":    nodeType,
	}).Observe(duration.Seconds())

	// Track result
	status := "success"
	if err != nil {
		status = "error"
		m.nodeErrorsTotal.With(prometheus.Labels{
			"workspace_id": workspaceID,
			"node_type":    nodeType,
			"error_type":   classifyError(err),
		}).Inc()
	}

	m.nodeExecutionsTotal.With(prometheus.Labels{
		"workspace_id": workspaceID,
		"node_type":    nodeType,
		"status":       status,
	}).Inc()

	if err != nil {
		return nil, err
	}
	return result.Output, nil
}

// classifyError categorizes errors for metrics
func classifyError(err error) string {
	errStr := err.Error()

	switch {
	case contains(errStr, "timeout"):
		return "timeout"
	case contains(errStr, "cancelled") || contains(errStr, "context canceled"):
		return "cancelled"
	case contains(errStr, "rate limit"):
		return "rate_limit"
	case contains(errStr, "authentication") || contains(errStr, "unauthorized"):
		return "auth"
	case contains(errStr, "not found") || contains(errStr, "404"):
		return "not_found"
	case contains(errStr, "connection") || contains(errStr, "network"):
		return "network"
	default:
		return "unknown"
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// WorkflowMetrics tracks workflow-level metrics
type WorkflowMetrics struct {
	executionsTotal *prometheus.CounterVec
	duration        *prometheus.HistogramVec
	nodesPerExec    *prometheus.HistogramVec
	activeWorkflows *prometheus.GaugeVec
}

// NewWorkflowMetrics creates workflow metrics
func NewWorkflowMetrics() *WorkflowMetrics {
	return &WorkflowMetrics{
		executionsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "linkflow_workflow_executions_total",
				Help: "Total number of workflow executions",
			},
			[]string{"workspace_id", "trigger_type", "status"},
		),
		duration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "linkflow_workflow_duration_seconds",
				Help:    "Duration of workflow executions in seconds",
				Buckets: prometheus.ExponentialBuckets(0.1, 2, 15), // 100ms to ~27min
			},
			[]string{"workspace_id"},
		),
		nodesPerExec: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "linkflow_workflow_nodes_count",
				Help:    "Number of nodes executed per workflow",
				Buckets: prometheus.LinearBuckets(1, 5, 20), // 1 to 100
			},
			[]string{"workspace_id"},
		),
		activeWorkflows: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "linkflow_active_workflows",
				Help: "Number of currently executing workflows",
			},
			[]string{"workspace_id"},
		),
	}
}

// RecordStart records workflow start
func (m *WorkflowMetrics) RecordStart(workspaceID string) {
	m.activeWorkflows.With(prometheus.Labels{"workspace_id": workspaceID}).Inc()
}

// RecordComplete records workflow completion
func (m *WorkflowMetrics) RecordComplete(workspaceID, triggerType, status string, duration time.Duration, nodesCount int) {
	m.activeWorkflows.With(prometheus.Labels{"workspace_id": workspaceID}).Dec()

	m.executionsTotal.With(prometheus.Labels{
		"workspace_id": workspaceID,
		"trigger_type": triggerType,
		"status":       status,
	}).Inc()

	m.duration.With(prometheus.Labels{
		"workspace_id": workspaceID,
	}).Observe(duration.Seconds())

	m.nodesPerExec.With(prometheus.Labels{
		"workspace_id": workspaceID,
	}).Observe(float64(nodesCount))
}

// MetricsCollector implements processor.MetricsCollector
type MetricsCollector struct {
	nodeMetrics     *MetricsMiddleware
	workflowMetrics *WorkflowMetrics
	mu              sync.RWMutex
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		nodeMetrics:     NewMetricsMiddleware(),
		workflowMetrics: NewWorkflowMetrics(),
	}
}

// RecordNodeExecution records node execution metrics
func (mc *MetricsCollector) RecordNodeExecution(workspaceID, nodeType string, duration time.Duration, err error) {
	status := "success"
	if err != nil {
		status = "error"
	}

	mc.nodeMetrics.nodeExecutionsTotal.With(prometheus.Labels{
		"workspace_id": workspaceID,
		"node_type":    nodeType,
		"status":       status,
	}).Inc()

	mc.nodeMetrics.nodeDuration.With(prometheus.Labels{
		"workspace_id": workspaceID,
		"node_type":    nodeType,
	}).Observe(duration.Seconds())

	if err != nil {
		mc.nodeMetrics.nodeErrorsTotal.With(prometheus.Labels{
			"workspace_id": workspaceID,
			"node_type":    nodeType,
			"error_type":   classifyError(err),
		}).Inc()
	}
}

// RecordWorkflowExecution records workflow execution metrics
func (mc *MetricsCollector) RecordWorkflowExecution(workspaceID string, duration time.Duration, nodesCount int, err error) {
	status := "success"
	if err != nil {
		status = "error"
	}

	mc.workflowMetrics.executionsTotal.With(prometheus.Labels{
		"workspace_id": workspaceID,
		"trigger_type": "unknown", // Would need to be passed in
		"status":       status,
	}).Inc()

	mc.workflowMetrics.duration.With(prometheus.Labels{
		"workspace_id": workspaceID,
	}).Observe(duration.Seconds())

	mc.workflowMetrics.nodesPerExec.With(prometheus.Labels{
		"workspace_id": workspaceID,
	}).Observe(float64(nodesCount))
}

// QueueMetrics tracks queue-related metrics
type QueueMetrics struct {
	queueDepth          *prometheus.GaugeVec
	queueLatency        *prometheus.HistogramVec
	processingTime      *prometheus.HistogramVec
	retriesTotal        *prometheus.CounterVec
	deadLetterTotal     *prometheus.CounterVec
}

// NewQueueMetrics creates queue metrics
func NewQueueMetrics() *QueueMetrics {
	return &QueueMetrics{
		queueDepth: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "linkflow_queue_depth",
				Help: "Current queue depth",
			},
			[]string{"queue_name"},
		),
		queueLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "linkflow_queue_latency_seconds",
				Help:    "Time from enqueue to dequeue",
				Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
			},
			[]string{"queue_name"},
		),
		processingTime: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "linkflow_queue_processing_seconds",
				Help:    "Time to process a queue item",
				Buckets: prometheus.ExponentialBuckets(0.01, 2, 15),
			},
			[]string{"queue_name"},
		),
		retriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "linkflow_queue_retries_total",
				Help: "Total number of queue retries",
			},
			[]string{"queue_name"},
		),
		deadLetterTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "linkflow_queue_dead_letter_total",
				Help: "Total items sent to dead letter queue",
			},
			[]string{"queue_name"},
		),
	}
}

// RecordEnqueue records when an item is enqueued
func (m *QueueMetrics) RecordEnqueue(queueName string) {
	m.queueDepth.With(prometheus.Labels{"queue_name": queueName}).Inc()
}

// RecordDequeue records when an item is dequeued
func (m *QueueMetrics) RecordDequeue(queueName string, latency time.Duration) {
	m.queueDepth.With(prometheus.Labels{"queue_name": queueName}).Dec()
	m.queueLatency.With(prometheus.Labels{"queue_name": queueName}).Observe(latency.Seconds())
}

// RecordProcessed records when an item is processed
func (m *QueueMetrics) RecordProcessed(queueName string, duration time.Duration) {
	m.processingTime.With(prometheus.Labels{"queue_name": queueName}).Observe(duration.Seconds())
}

// RecordRetry records a retry
func (m *QueueMetrics) RecordRetry(queueName string) {
	m.retriesTotal.With(prometheus.Labels{"queue_name": queueName}).Inc()
}

// RecordDeadLetter records a dead letter
func (m *QueueMetrics) RecordDeadLetter(queueName string) {
	m.deadLetterTotal.With(prometheus.Labels{"queue_name": queueName}).Inc()
}
