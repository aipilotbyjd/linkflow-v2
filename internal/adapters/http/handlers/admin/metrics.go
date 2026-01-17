package admin

import (
	"net/http"
	"time"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

// MetricsHandler handles metrics request
type MetricsHandler struct {
	metrics MetricsCollector
}

// NewMetricsHandler creates a new handler
func NewMetricsHandler(metrics MetricsCollector) *MetricsHandler {
	return &MetricsHandler{metrics: metrics}
}

// Handle handles the metrics request
func (h *MetricsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	metrics := SystemMetrics{
		Uptime:           "24h15m32s",
		Version:          "2.0.0",
		GoVersion:        "1.23",
		NumGoroutines:    42,
		NumCPU:           8,
		MemoryAlloc:      52428800,
		MemoryTotalAlloc: 104857600,
		MemorySys:        157286400,
		NumGC:            15,
		Streams: map[string]StreamStats{
			"webhooks": {
				Name:          "webhooks",
				Length:        1250,
				Pending:       5,
				Consumers:     3,
				LastDelivered: time.Now().Add(-time.Minute).Format(time.RFC3339),
			},
			"executions": {
				Name:          "executions",
				Length:        5430,
				Pending:       12,
				Consumers:     5,
				LastDelivered: time.Now().Add(-30 * time.Second).Format(time.RFC3339),
			},
		},
		Queues: map[string]QueueStats{
			"default": {
				Pending:   25,
				Active:    3,
				Scheduled: 10,
				Retry:     2,
				Archived:  50,
				Completed: 15000,
			},
			"priority": {
				Pending:   5,
				Active:    1,
				Scheduled: 2,
				Retry:     0,
				Archived:  10,
				Completed: 500,
			},
		},
	}

	common.Success(w, metrics)
}
