package schedule

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	scheduleCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/schedule"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/validation"
)

// CreateHandler handles schedule creation
type CreateHandler struct {
	handler *scheduleCmd.CreateScheduleHandler
}

// NewCreateHandler creates a new handler
func NewCreateHandler(handler *scheduleCmd.CreateScheduleHandler) *CreateHandler {
	return &CreateHandler{handler: handler}
}

// Handle handles the create schedule request
func (h *CreateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
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

	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		common.BadRequest(w, "workspace context required")
		return
	}

	userClaims := middleware.GetUserFromContext(r.Context())
	if userClaims == nil {
		common.Unauthorized(w, "")
		return
	}

	timezone := req.Timezone
	if timezone == "" {
		timezone = "UTC"
	}

	sched, err := h.handler.Handle(r.Context(), scheduleCmd.CreateScheduleCommand{
		WorkflowID:     req.WorkflowID,
		WorkspaceID:    wsCtx.WorkspaceID,
		CreatedBy:      userClaims.UserID,
		Name:           req.Name,
		Description:    req.Description,
		CronExpression: req.CronExpression,
		Timezone:       timezone,
		InputData:      req.InputData,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Created(w, ToScheduleResponse(sched))
}
