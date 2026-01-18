package analytics

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
)

type GetWorkspaceAnalyticsQuery struct {
	WorkspaceID uuid.UUID
	StartDate   time.Time
	EndDate     time.Time
}

type WorkspaceAnalyticsResult struct {
	Stats      *execution.Stats     `json:"stats"`
	DailyStats []execution.DailyStat `json:"daily_stats"`
}

type GetWorkspaceAnalyticsHandler struct {
	statsRepo execution.StatsRepository
}

func NewGetWorkspaceAnalyticsHandler(statsRepo execution.StatsRepository) *GetWorkspaceAnalyticsHandler {
	return &GetWorkspaceAnalyticsHandler{statsRepo: statsRepo}
}

func (h *GetWorkspaceAnalyticsHandler) Handle(ctx context.Context, q GetWorkspaceAnalyticsQuery) (*WorkspaceAnalyticsResult, error) {
	stats, err := h.statsRepo.GetStats(ctx, q.WorkspaceID, q.StartDate, q.EndDate)
	if err != nil {
		return nil, err
	}

	days := int(q.EndDate.Sub(q.StartDate).Hours() / 24)
	if days <= 0 {
		days = 30
	}

	dailyStats, err := h.statsRepo.GetDailyStats(ctx, q.WorkspaceID, days)
	if err != nil {
		return nil, err
	}

	return &WorkspaceAnalyticsResult{
		Stats:      stats,
		DailyStats: dailyStats,
	}, nil
}
