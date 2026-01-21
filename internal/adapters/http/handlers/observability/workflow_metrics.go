package observability

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/observability"
)

// WorkflowMetricsHandler handles workflow metrics requests
type WorkflowMetricsHandler struct{}

func NewWorkflowMetricsHandler() *WorkflowMetricsHandler {
	return &WorkflowMetricsHandler{}
}

func (h *WorkflowMetricsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workflowID, err := uuid.Parse(chi.URLParam(r, "workflowId"))
	if err != nil {
		common.BadRequest(w, "Invalid workflow ID")
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "day"
	}

	// In production, fetch from repository
	// For now, return placeholder metrics
	metrics := observability.WorkflowMetrics{
		WorkflowID:           workflowID,
		WorkflowName:         "Sample Workflow",
		Period:               period,
		TotalExecutions:      1250,
		SuccessfulExecutions: 1180,
		FailedExecutions:     70,
		SuccessRate:          94.4,
		ErrorRate:            5.6,
		AvgDuration:          2340,
		MinDuration:          450,
		MaxDuration:          15000,
		P50Duration:          1800,
		P95Duration:          5200,
		P99Duration:          12000,
		AvgNodesExecuted:     8.5,
		TotalCreditsUsed:     125.5,
		AvgCreditsPerRun:     0.1,
		EstimatedCostUSD:     12.55,
		StartTime:            time.Now().Add(-24 * time.Hour),
		EndTime:              time.Now(),
	}

	nodeMetrics := []observability.NodeMetrics{
		{NodeID: "node_1", NodeType: "trigger.webhook", NodeName: "Webhook Trigger", Executions: 1250, Failures: 0, AvgDuration: 5, ErrorRate: 0},
		{NodeID: "node_2", NodeType: "action.http", NodeName: "API Call", Executions: 1250, Failures: 50, AvgDuration: 1200, ErrorRate: 4.0},
		{NodeID: "node_3", NodeType: "logic.if", NodeName: "Check Status", Executions: 1200, Failures: 5, AvgDuration: 2, ErrorRate: 0.4},
		{NodeID: "node_4", NodeType: "integration.slack", NodeName: "Send Notification", Executions: 800, Failures: 15, AvgDuration: 350, ErrorRate: 1.9},
	}

	common.Success(w, WorkflowMetricsResponse{
		Metrics:     metrics,
		NodeMetrics: nodeMetrics,
	})
}
