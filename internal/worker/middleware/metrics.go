package middleware

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/linkflow-ai/linkflow/internal/worker/processor"
)

// MetricsMiddleware collects metrics for node execution
type MetricsMiddleware struct {
	nodeExecutions sync.Map // key -> *nodeMetrics
}

type nodeMetrics struct {
	total    int64
	errors   int64
	duration int64 // total duration in ms
}

// NewMetricsMiddleware creates a new metrics middleware
func NewMetricsMiddleware() *MetricsMiddleware {
	return &MetricsMiddleware{}
}

// Execute implements Middleware
func (m *MetricsMiddleware) Execute(ctx context.Context, rctx *processor.RuntimeContext, node *processor.NodeDefinition, next NextFunc) (map[string]interface{}, error) {
	key := node.Type

	// Track duration
	startTime := time.Now()
	result, err := next(ctx)
	duration := time.Since(startTime).Milliseconds()

	// Get or create metrics
	metricsI, _ := m.nodeExecutions.LoadOrStore(key, &nodeMetrics{})
	metrics := metricsI.(*nodeMetrics)

	atomic.AddInt64(&metrics.total, 1)
	atomic.AddInt64(&metrics.duration, duration)

	if err != nil {
		atomic.AddInt64(&metrics.errors, 1)
		return nil, err
	}

	return result.Output, nil
}

// GetMetrics returns collected metrics
func (m *MetricsMiddleware) GetMetrics() map[string]interface{} {
	result := make(map[string]interface{})

	m.nodeExecutions.Range(func(key, value interface{}) bool {
		metrics := value.(*nodeMetrics)
		total := atomic.LoadInt64(&metrics.total)
		errors := atomic.LoadInt64(&metrics.errors)
		duration := atomic.LoadInt64(&metrics.duration)

		avgDuration := int64(0)
		if total > 0 {
			avgDuration = duration / total
		}

		result[key.(string)] = map[string]interface{}{
			"total":            total,
			"errors":           errors,
			"error_rate":       float64(errors) / float64(total),
			"avg_duration_ms":  avgDuration,
			"total_duration_ms": duration,
		}
		return true
	})

	return result
}

// Reset resets all metrics
func (m *MetricsMiddleware) Reset() {
	m.nodeExecutions = sync.Map{}
}

// MetricsCollector implements processor.MetricsCollector
type MetricsCollector struct {
	nodeMetrics     *MetricsMiddleware
	workflowMetrics *WorkflowMetrics
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
	key := nodeType

	metricsI, _ := mc.nodeMetrics.nodeExecutions.LoadOrStore(key, &nodeMetrics{})
	metrics := metricsI.(*nodeMetrics)

	atomic.AddInt64(&metrics.total, 1)
	atomic.AddInt64(&metrics.duration, duration.Milliseconds())

	if err != nil {
		atomic.AddInt64(&metrics.errors, 1)
	}
}

// RecordWorkflowExecution records workflow execution metrics
func (mc *MetricsCollector) RecordWorkflowExecution(workspaceID string, duration time.Duration, nodesCount int, err error) {
	mc.workflowMetrics.RecordExecution(duration, nodesCount, err)
}

// GetMetrics returns all metrics
func (mc *MetricsCollector) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"nodes":     mc.nodeMetrics.GetMetrics(),
		"workflows": mc.workflowMetrics.GetMetrics(),
	}
}

// WorkflowMetrics tracks workflow-level metrics
type WorkflowMetrics struct {
	total      int64
	completed  int64
	failed     int64
	duration   int64
	totalNodes int64
}

// NewWorkflowMetrics creates workflow metrics
func NewWorkflowMetrics() *WorkflowMetrics {
	return &WorkflowMetrics{}
}

// RecordExecution records a workflow execution
func (m *WorkflowMetrics) RecordExecution(duration time.Duration, nodesCount int, err error) {
	atomic.AddInt64(&m.total, 1)
	atomic.AddInt64(&m.duration, duration.Milliseconds())
	atomic.AddInt64(&m.totalNodes, int64(nodesCount))

	if err != nil {
		atomic.AddInt64(&m.failed, 1)
	} else {
		atomic.AddInt64(&m.completed, 1)
	}
}

// GetMetrics returns workflow metrics
func (m *WorkflowMetrics) GetMetrics() map[string]interface{} {
	total := atomic.LoadInt64(&m.total)
	completed := atomic.LoadInt64(&m.completed)
	failed := atomic.LoadInt64(&m.failed)
	duration := atomic.LoadInt64(&m.duration)
	totalNodes := atomic.LoadInt64(&m.totalNodes)

	avgDuration := int64(0)
	avgNodes := float64(0)
	if total > 0 {
		avgDuration = duration / total
		avgNodes = float64(totalNodes) / float64(total)
	}

	return map[string]interface{}{
		"total":             total,
		"completed":         completed,
		"failed":            failed,
		"success_rate":      float64(completed) / float64(total),
		"avg_duration_ms":   avgDuration,
		"total_duration_ms": duration,
		"avg_nodes":         avgNodes,
	}
}

// QueueMetrics tracks queue-related metrics
type QueueMetrics struct {
	enqueued   int64
	dequeued   int64
	processed  int64
	retries    int64
	deadLetter int64
}

// NewQueueMetrics creates queue metrics
func NewQueueMetrics() *QueueMetrics {
	return &QueueMetrics{}
}

// RecordEnqueue records when an item is enqueued
func (m *QueueMetrics) RecordEnqueue() {
	atomic.AddInt64(&m.enqueued, 1)
}

// RecordDequeue records when an item is dequeued
func (m *QueueMetrics) RecordDequeue() {
	atomic.AddInt64(&m.dequeued, 1)
}

// RecordProcessed records when an item is processed
func (m *QueueMetrics) RecordProcessed() {
	atomic.AddInt64(&m.processed, 1)
}

// RecordRetry records a retry
func (m *QueueMetrics) RecordRetry() {
	atomic.AddInt64(&m.retries, 1)
}

// RecordDeadLetter records a dead letter
func (m *QueueMetrics) RecordDeadLetter() {
	atomic.AddInt64(&m.deadLetter, 1)
}

// GetMetrics returns queue metrics
func (m *QueueMetrics) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"enqueued":    atomic.LoadInt64(&m.enqueued),
		"dequeued":    atomic.LoadInt64(&m.dequeued),
		"processed":   atomic.LoadInt64(&m.processed),
		"retries":     atomic.LoadInt64(&m.retries),
		"dead_letter": atomic.LoadInt64(&m.deadLetter),
	}
}
