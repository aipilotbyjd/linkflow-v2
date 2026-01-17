package execution

import (
	"net/http"
	"time"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
)

type GetExecutionStatsHandler struct {
	executionRepo execution.Repository
}

func NewGetExecutionStatsHandler(executionRepo execution.Repository) *GetExecutionStatsHandler {
	return &GetExecutionStatsHandler{executionRepo: executionRepo}
}

func (h *GetExecutionStatsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		common.BadRequest(w, "workspace context required")
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "24h"
	}

	var duration time.Duration
	switch period {
	case "1h":
		duration = time.Hour
	case "24h":
		duration = 24 * time.Hour
	case "7d":
		duration = 7 * 24 * time.Hour
	case "30d":
		duration = 30 * 24 * time.Hour
	default:
		duration = 24 * time.Hour
	}

	now := time.Now()
	startDate := now.Add(-duration)

	executions, total, err := h.executionRepo.FindByWorkspaceID(r.Context(), wsCtx.WorkspaceID, nil)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	// Calculate stats
	byStatus := make(map[string]int64)
	var totalDuration int64
	var countWithDuration int64
	var filteredCount int64

	for _, exec := range executions {
		if exec.StartedAt != nil && exec.StartedAt.After(startDate) {
			filteredCount++
			byStatus[string(exec.Status)]++
			if exec.CompletedAt != nil {
				totalDuration += exec.CompletedAt.Sub(*exec.StartedAt).Milliseconds()
				countWithDuration++
			}
		}
	}

	var avgDuration int64
	if countWithDuration > 0 {
		avgDuration = totalDuration / countWithDuration
	}

	response := ExecutionStatsResponse{
		Total:         total,
		ByStatus:      byStatus,
		AvgDurationMs: avgDuration,
		Period:        period,
		StartDate:     startDate.Format(time.RFC3339),
		EndDate:       now.Format(time.RFC3339),
	}

	common.Success(w, response)
}
