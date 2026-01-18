package analytics

import (
	"time"

	"github.com/linkflow-ai/linkflow/internal/core/application/query/analytics"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
)

// Response DTOs

type StatsResponse struct {
	Total       int64   `json:"total"`
	Completed   int64   `json:"completed"`
	Failed      int64   `json:"failed"`
	Canceled    int64   `json:"canceled"`
	Running     int64   `json:"running"`
	Queued      int64   `json:"queued"`
	AvgDuration int64   `json:"avgDurationMs"`
	SuccessRate float64 `json:"successRate"`
}

type DailyStatResponse struct {
	Date      string `json:"date"`
	Total     int64  `json:"total"`
	Completed int64  `json:"completed"`
	Failed    int64  `json:"failed"`
}

type WorkspaceAnalyticsResponse struct {
	Stats      StatsResponse       `json:"stats"`
	DailyStats []DailyStatResponse `json:"dailyStats"`
}

// Mappers

func ToStatsResponse(s *execution.Stats) StatsResponse {
	if s == nil {
		return StatsResponse{}
	}
	return StatsResponse{
		Total:       s.Total,
		Completed:   s.Completed,
		Failed:      s.Failed,
		Canceled:    s.Canceled,
		Running:     s.Running,
		Queued:      s.Queued,
		AvgDuration: s.AvgDuration.Milliseconds(),
		SuccessRate: s.SuccessRate,
	}
}

func ToDailyStatResponse(d execution.DailyStat) DailyStatResponse {
	return DailyStatResponse{
		Date:      d.Date.Format(time.DateOnly),
		Total:     d.Total,
		Completed: d.Completed,
		Failed:    d.Failed,
	}
}

func ToDailyStatResponses(stats []execution.DailyStat) []DailyStatResponse {
	responses := make([]DailyStatResponse, len(stats))
	for i, s := range stats {
		responses[i] = ToDailyStatResponse(s)
	}
	return responses
}

func ToWorkspaceAnalyticsResponse(r *analytics.WorkspaceAnalyticsResult) WorkspaceAnalyticsResponse {
	return WorkspaceAnalyticsResponse{
		Stats:      ToStatsResponse(r.Stats),
		DailyStats: ToDailyStatResponses(r.DailyStats),
	}
}
