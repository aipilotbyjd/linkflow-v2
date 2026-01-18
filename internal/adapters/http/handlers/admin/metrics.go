package admin

import (
	"net/http"

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
	metricsData := h.metrics.CollectMetrics()
	common.Success(w, metricsData)
}
