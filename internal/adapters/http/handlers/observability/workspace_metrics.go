package observability

import (
	"net/http"
	"time"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/observability"
)

// WorkspaceMetricsHandler handles workspace metrics requests
type WorkspaceMetricsHandler struct{}

func NewWorkspaceMetricsHandler() *WorkspaceMetricsHandler {
	return &WorkspaceMetricsHandler{}
}

func (h *WorkspaceMetricsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workspaceID := middleware.GetWorkspaceID(r.Context())

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "day"
	}

	// In production, fetch from repository
	metrics := observability.WorkspaceMetrics{
		WorkspaceID:          workspaceID,
		WorkspaceName:        "My Workspace",
		Period:               period,
		TotalWorkflows:       45,
		ActiveWorkflows:      28,
		TotalExecutions:      15420,
		SuccessfulExecutions: 14650,
		FailedExecutions:     770,
		SuccessRate:          95.0,
		AvgDuration:          3200,
		TopWorkflowsByExecutions: []observability.WorkflowSummary{
			{WorkflowName: "Order Processing", Value: 5200},
			{WorkflowName: "User Onboarding", Value: 3100},
			{WorkflowName: "Daily Reports", Value: 2800},
		},
		TopWorkflowsByErrors: []observability.WorkflowSummary{
			{WorkflowName: "API Sync", Value: 120, Label: "120 failures"},
			{WorkflowName: "Payment Webhook", Value: 85, Label: "85 failures"},
		},
		TotalCreditsUsed:   1542.5,
		CreditsRemaining:   8457.5,
		EstimatedCostUSD:   154.25,
		APICallsCount:      45000,
		DataProcessedBytes: 125000000,
		ExecutionTrend:     generateTrend(period),
		StartTime:          time.Now().Add(-24 * time.Hour),
		EndTime:            time.Now(),
	}

	heatmap := generateHeatmap()

	common.Success(w, WorkspaceMetricsResponse{
		Metrics: metrics,
		Heatmap: &heatmap,
	})
}

func generateTrend(period string) []observability.TrendPoint {
	var points []observability.TrendPoint
	now := time.Now()

	count := 24
	interval := time.Hour
	if period == "week" {
		count = 7
		interval = 24 * time.Hour
	} else if period == "month" {
		count = 30
		interval = 24 * time.Hour
	}

	for i := count - 1; i >= 0; i-- {
		t := now.Add(-time.Duration(i) * interval)
		points = append(points, observability.TrendPoint{
			Timestamp: t,
			Value:     float64(500 + (i%5)*100),
		})
	}
	return points
}

func generateHeatmap() observability.ExecutionHeatmap {
	var cells []observability.HeatmapCell
	for day := 0; day < 7; day++ {
		for hour := 0; hour < 24; hour++ {
			count := int64(10 + (hour%12)*5 + day*3)
			cells = append(cells, observability.HeatmapCell{
				DayOfWeek:   day,
				HourOfDay:   hour,
				Count:       count,
				AvgDuration: float64(1500 + hour*50),
			})
		}
	}
	return observability.ExecutionHeatmap{Data: cells}
}
