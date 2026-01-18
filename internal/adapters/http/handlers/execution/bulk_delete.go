package execution

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/validation"
)

type BulkDeleteExecutionsHandler struct {
	executionRepo execution.Repository
}

func NewBulkDeleteExecutionsHandler(executionRepo execution.Repository) *BulkDeleteExecutionsHandler {
	return &BulkDeleteExecutionsHandler{executionRepo: executionRepo}
}

func (h *BulkDeleteExecutionsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		common.BadRequest(w, "workspace context required")
		return
	}

	var req BulkDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}

	if errors := validation.Validate(req); len(errors) > 0 {
		details := make([]common.ValidationDetail, len(errors))
		for i, e := range errors {
			details[i] = common.ValidationDetail{Field: e.Field, Message: e.Message}
		}
		common.ValidationErrors(w, details)
		return
	}

	var deletedCount int64

	if len(req.ExecutionIDs) > 0 {
		// Delete specific executions
		for _, idStr := range req.ExecutionIDs {
			id, err := uuid.Parse(idStr)
			if err != nil {
				continue
			}

			// Verify execution belongs to workspace
			exec, err := h.executionRepo.FindByID(r.Context(), id)
			if err != nil || exec.WorkspaceID != wsCtx.WorkspaceID {
				continue
			}

			if err := h.executionRepo.Delete(r.Context(), id); err == nil {
				deletedCount++
			}
		}
	} else if req.OlderThan != nil {
		// Delete by age
		olderThan, err := time.Parse(time.RFC3339, *req.OlderThan)
		if err != nil {
			common.BadRequest(w, "invalid older_than format, use RFC3339")
			return
		}

		executions, _, _ := h.executionRepo.FindByWorkspaceID(r.Context(), wsCtx.WorkspaceID, nil)
		for _, exec := range executions {
			if exec.StartedAt.Before(olderThan) {
				// Check status filter
				if req.Status != nil && string(exec.Status) != *req.Status {
					continue
				}
				if err := h.executionRepo.Delete(r.Context(), exec.ID); err == nil {
					deletedCount++
				}
			}
		}
	} else {
		common.BadRequest(w, "either execution_ids or older_than is required")
		return
	}

	common.Success(w, BulkDeleteResponse{Deleted: deletedCount})
}
