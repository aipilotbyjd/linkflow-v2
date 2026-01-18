package workflow

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	workflowCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/workflow"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/validation"
)

// CreateHandler handles workflow creation
type CreateHandler struct {
	handler *workflowCmd.CreateWorkflowHandler
}

// NewCreateHandler creates a new create handler
func NewCreateHandler(handler *workflowCmd.CreateWorkflowHandler) *CreateHandler {
	return &CreateHandler{handler: handler}
}

// Handle handles the create workflow request
func (h *CreateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req CreateWorkflowRequest
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

	wf, err := h.handler.Handle(r.Context(), workflowCmd.CreateWorkflowCommand{
		WorkspaceID: wsCtx.WorkspaceID,
		CreatedBy:   userClaims.UserID,
		Name:        req.Name,
		Description: req.Description,
		Nodes:       req.Nodes,
		Connections: req.Connections,
		Settings:    req.Settings,
		Tags:        req.Tags,
		Color:       req.Color,
		Icon:        req.Icon,
		Category:    req.Category,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Created(w, ToWorkflowResponse(wf))
}
