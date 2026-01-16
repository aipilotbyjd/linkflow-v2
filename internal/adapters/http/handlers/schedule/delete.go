package schedule

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/schedule"
)

// DeleteHandler handles schedule deletion
type DeleteHandler struct {
	repo schedule.Repository
}

// NewDeleteHandler creates a new handler
func NewDeleteHandler(repo schedule.Repository) *DeleteHandler {
	return &DeleteHandler{repo: repo}
}

// Handle handles the delete schedule request
func (h *DeleteHandler) Handle(w http.ResponseWriter, r *http.Request) {
	scheduleIDStr := chi.URLParam(r, "scheduleId")
	scheduleID, err := uuid.Parse(scheduleIDStr)
	if err != nil {
		common.BadRequest(w, "invalid schedule ID")
		return
	}

	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		common.BadRequest(w, "workspace context required")
		return
	}

	// Get existing schedule
	sched, err := h.repo.FindByID(r.Context(), scheduleID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	// Verify workspace ownership
	if sched.WorkspaceID != wsCtx.WorkspaceID {
		common.NotFound(w, "schedule not found")
		return
	}

	if err := h.repo.Delete(r.Context(), scheduleID); err != nil {
		common.HandleError(w, err)
		return
	}

	common.NoContent(w)
}
