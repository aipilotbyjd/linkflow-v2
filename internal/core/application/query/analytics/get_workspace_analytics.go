package analytics

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
)

type GetWorkspaceAnalyticsQuery struct {
	WorkspaceID uuid.UUID
	StartDate   time.Time
	EndDate     time.Time
}

type WorkspaceAnalyticsResult struct {
	TotalWorkflows    int64              `json:"total_workflows"`
	ActiveWorkflows   int64              `json:"active_workflows"`
	TotalExecutions   int64              `json:"total_executions"`
	SuccessfulRuns    int64              `json:"successful_runs"`
	FailedRuns        int64              `json:"failed_runs"`
	AverageDurationMs float64            `json:"average_duration_ms"`
	ExecutionsByDay   map[string]int64   `json:"executions_by_day"`
	TopWorkflows      []WorkflowStats    `json:"top_workflows"`
}

type WorkflowStats struct {
	WorkflowID   uuid.UUID `json:"workflow_id"`
	WorkflowName string    `json:"workflow_name"`
	Executions   int64     `json:"executions"`
	SuccessRate  float64   `json:"success_rate"`
}

type GetWorkspaceAnalyticsHandler struct {
	workflowRepo workflow.Repository
	statsRepo    execution.StatsRepository
}

func NewGetWorkspaceAnalyticsHandler(workflowRepo workflow.Repository, statsRepo execution.StatsRepository) *GetWorkspaceAnalyticsHandler {
	return &GetWorkspaceAnalyticsHandler{workflowRepo: workflowRepo, statsRepo: statsRepo}
}

func (h *GetWorkspaceAnalyticsHandler) Handle(ctx context.Context, q GetWorkspaceAnalyticsQuery) (*WorkspaceAnalyticsResult, error) {
	// Get workflow counts
	workflows, total, err := h.workflowRepo.FindByWorkspaceID(ctx, q.WorkspaceID, nil)
	if err != nil {
		return nil, err
	}

	var activeCount int64
	for _, wf := range workflows {
		if wf.Status == workflow.StatusActive {
			activeCount++
		}
	}

	// Get execution stats
	execStats, err := h.statsRepo.GetStats(ctx, q.WorkspaceID, q.StartDate, q.EndDate)
	if err != nil {
		return nil, err
	}

	return &WorkspaceAnalyticsResult{
		TotalWorkflows:    total,
		ActiveWorkflows:   activeCount,
		TotalExecutions:   execStats.Total,
		SuccessfulRuns:    execStats.Completed,
		FailedRuns:        execStats.Failed,
		AverageDurationMs: float64(execStats.AvgDuration.Milliseconds()),
		ExecutionsByDay:   make(map[string]int64),
		TopWorkflows:      []WorkflowStats{},
	}, nil
}
