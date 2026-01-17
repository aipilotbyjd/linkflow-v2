package schedule

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/schedule"
)

// ResumeHandler handles schedule resumption
type ResumeHandler struct {
	repo schedule.Repository
}

// NewResumeHandler creates a new handler
func NewResumeHandler(repo schedule.Repository) *ResumeHandler {
	return &ResumeHandler{repo: repo}
}

// Handle handles the resume schedule request
func (h *ResumeHandler) Handle(w http.ResponseWriter, r *http.Request) {
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

	// Resume the schedule (this also recalculates next run time)
	sched.Resume()

	if err := h.repo.Update(r.Context(), sched); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, ToScheduleResponse(sched))
}
