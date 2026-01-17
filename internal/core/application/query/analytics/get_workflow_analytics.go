package analytics

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
)

type GetWorkflowAnalyticsQuery struct {
	WorkflowID uuid.UUID
	StartDate  time.Time
	EndDate    time.Time
}

type WorkflowAnalyticsResult struct {
	WorkflowID        uuid.UUID        `json:"workflow_id"`
	TotalExecutions   int64            `json:"total_executions"`
	SuccessfulRuns    int64            `json:"successful_runs"`
	FailedRuns        int64            `json:"failed_runs"`
	PendingRuns       int64            `json:"pending_runs"`
	AverageDurationMs float64          `json:"average_duration_ms"`
	MinDurationMs     int64            `json:"min_duration_ms"`
	MaxDurationMs     int64            `json:"max_duration_ms"`
	SuccessRate       float64          `json:"success_rate"`
	ExecutionsByDay   map[string]int64 `json:"executions_by_day"`
	ExecutionsByHour  map[int]int64    `json:"executions_by_hour"`
}

type GetWorkflowAnalyticsHandler struct {
	statsRepo execution.StatsRepository
}

func NewGetWorkflowAnalyticsHandler(statsRepo execution.StatsRepository) *GetWorkflowAnalyticsHandler {
	return &GetWorkflowAnalyticsHandler{statsRepo: statsRepo}
}

func (h *GetWorkflowAnalyticsHandler) Handle(ctx context.Context, q GetWorkflowAnalyticsQuery) (*WorkflowAnalyticsResult, error) {
	stats, err := h.statsRepo.GetWorkflowStats(ctx, q.WorkflowID, q.StartDate, q.EndDate)
	if err != nil {
		return nil, err
	}

	return &WorkflowAnalyticsResult{
		WorkflowID:        q.WorkflowID,
		TotalExecutions:   stats.Total,
		SuccessfulRuns:    stats.Completed,
		FailedRuns:        stats.Failed,
		PendingRuns:       stats.Queued,
		AverageDurationMs: float64(stats.AvgDuration.Milliseconds()),
		MinDurationMs:     0,
		MaxDurationMs:     0,
		SuccessRate:       stats.SuccessRate,
		ExecutionsByDay:   make(map[string]int64),
		ExecutionsByHour:  make(map[int]int64),
	}, nil
}
