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

type GetWorkflowAnalyticsHandler struct {
	statsRepo execution.StatsRepository
}

func NewGetWorkflowAnalyticsHandler(statsRepo execution.StatsRepository) *GetWorkflowAnalyticsHandler {
	return &GetWorkflowAnalyticsHandler{statsRepo: statsRepo}
}

func (h *GetWorkflowAnalyticsHandler) Handle(ctx context.Context, q GetWorkflowAnalyticsQuery) (*execution.Stats, error) {
	return h.statsRepo.GetWorkflowStats(ctx, q.WorkflowID, q.StartDate, q.EndDate)
}
