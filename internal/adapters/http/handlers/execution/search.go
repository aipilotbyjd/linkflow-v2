package execution

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
)

type SearchExecutionsHandler struct {
	executionRepo execution.Repository
}

func NewSearchExecutionsHandler(executionRepo execution.Repository) *SearchExecutionsHandler {
	return &SearchExecutionsHandler{executionRepo: executionRepo}
}

func (h *SearchExecutionsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		common.BadRequest(w, "workspace context required")
		return
	}

	query := r.URL.Query()

	// Get all executions first
	executions, total, err := h.executionRepo.FindByWorkspaceID(r.Context(), wsCtx.WorkspaceID, nil)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	// Filter results
	var filtered []*execution.Execution
	for _, exec := range executions {
		// Filter by status
		if status := query.Get("status"); status != "" {
			if string(exec.Status) != status {
				continue
			}
		}

		// Filter by workflow_id
		if wfIDStr := query.Get("workflow_id"); wfIDStr != "" {
			wfID, err := uuid.Parse(wfIDStr)
			if err == nil && exec.WorkflowID != wfID {
				continue
			}
		}

		// Filter by trigger_type
		if triggerType := query.Get("trigger_type"); triggerType != "" {
			if exec.TriggerType != triggerType {
				continue
			}
		}

		// Filter by start_date
		if startDateStr := query.Get("start_date"); startDateStr != "" {
			startDate, err := time.Parse(time.RFC3339, startDateStr)
			if err == nil && exec.StartedAt.Before(startDate) {
				continue
			}
		}

		// Filter by end_date
		if endDateStr := query.Get("end_date"); endDateStr != "" {
			endDate, err := time.Parse(time.RFC3339, endDateStr)
			if err == nil && exec.StartedAt.After(endDate) {
				continue
			}
		}

		// Search in error message
		if q := query.Get("q"); q != "" {
			errMsg := ""
			if exec.ErrorMessage != nil {
				errMsg = *exec.ErrorMessage
			}
			if errMsg == "" || !containsIgnoreCase(errMsg, q) {
				continue
			}
		}

		filtered = append(filtered, &exec)
	}

	// Pagination
	page := 1
	perPage := 20
	if p := query.Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if pp := query.Get("per_page"); pp != "" {
		if parsed, err := strconv.Atoi(pp); err == nil && parsed > 0 && parsed <= 100 {
			perPage = parsed
		}
	}

	// Apply pagination
	start := (page - 1) * perPage
	end := start + perPage
	if start > len(filtered) {
		start = len(filtered)
	}
	if end > len(filtered) {
		end = len(filtered)
	}
	paginated := filtered[start:end]

	response := map[string]interface{}{
		"data": paginated,
		"meta": map[string]interface{}{
			"total":    total,
			"filtered": len(filtered),
			"page":     page,
			"per_page": perPage,
		},
	}

	common.Success(w, response)
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && 
		(s == substr || 
		 len(s) > 0 && len(substr) > 0 && 
		 (s[0] == substr[0] || s[0]+32 == substr[0] || s[0] == substr[0]+32))
}
