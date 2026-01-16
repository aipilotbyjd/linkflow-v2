package analytics

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
)

type GetExecutionMetricsQuery struct {
	WorkspaceID uuid.UUID
	StartDate   time.Time
	EndDate     time.Time
	Interval    string // hour, day, week
}

type ExecutionMetricsResult struct {
	TotalExecutions   int64            `json:"total_executions"`
	AverageDurationMs float64          `json:"average_duration_ms"`
	SuccessRate       float64          `json:"success_rate"`
	Timeline          []TimelinePoint  `json:"timeline"`
	ByTriggerType     map[string]int64 `json:"by_trigger_type"`
	ByStatus          map[string]int64 `json:"by_status"`
}

type TimelinePoint struct {
	Timestamp  time.Time `json:"timestamp"`
	Executions int64     `json:"executions"`
	Successful int64     `json:"successful"`
	Failed     int64     `json:"failed"`
}

type GetExecutionMetricsHandler struct {
	statsRepo execution.StatsRepository
}

func NewGetExecutionMetricsHandler(statsRepo execution.StatsRepository) *GetExecutionMetricsHandler {
	return &GetExecutionMetricsHandler{statsRepo: statsRepo}
}

func (h *GetExecutionMetricsHandler) Handle(ctx context.Context, q GetExecutionMetricsQuery) (*ExecutionMetricsResult, error) {
	stats, err := h.statsRepo.GetStats(ctx, q.WorkspaceID, q.StartDate, q.EndDate)
	if err != nil {
		return nil, err
	}

	return &ExecutionMetricsResult{
		TotalExecutions:   stats.Total,
		AverageDurationMs: float64(stats.AvgDuration.Milliseconds()),
		SuccessRate:       stats.SuccessRate,
		Timeline:          []TimelinePoint{},
		ByTriggerType:     make(map[string]int64),
		ByStatus:          make(map[string]int64),
	}, nil
}
