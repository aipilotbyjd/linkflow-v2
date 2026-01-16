package schedule

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/schedule"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// UpdateRequest represents schedule update request
type UpdateRequest struct {
	Name           string     `json:"name" validate:"required"`
	Description    *string    `json:"description,omitempty"`
	CronExpression string     `json:"cron_expression" validate:"required"`
	Timezone       string     `json:"timezone"`
	InputData      types.JSON `json:"input_data,omitempty"`
}

// UpdateHandler handles schedule updates
type UpdateHandler struct {
	repo schedule.Repository
}

// NewUpdateHandler creates a new handler
func NewUpdateHandler(repo schedule.Repository) *UpdateHandler {
	return &UpdateHandler{repo: repo}
}

// Handle handles the update schedule request
func (h *UpdateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	scheduleIDStr := chi.URLParam(r, "scheduleId")
	scheduleID, err := uuid.Parse(scheduleIDStr)
	if err != nil {
		common.BadRequest(w, "invalid schedule ID")
		return
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "invalid request body")
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

	// Update schedule
	timezone := req.Timezone
	if timezone == "" {
		timezone = sched.Timezone
	}

	sched.Update(req.Name, req.Description, req.CronExpression, timezone)
	if req.InputData != nil {
		sched.UpdateInputData(req.InputData)
	}

	// Recalculate next run if still active
	if sched.IsActive {
		if nextRun, err := sched.CalculateNextRun(); err == nil {
			sched.SetNextRunAt(nextRun)
		}
	}

	if err := h.repo.Update(r.Context(), sched); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, toScheduleResponse(sched))
}
