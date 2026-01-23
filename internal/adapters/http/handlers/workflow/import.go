package workflow

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/validation"
)

type ImportHandler struct {
	workflowRepo workflow.Repository
}

func NewImportHandler(workflowRepo workflow.Repository) *ImportHandler {
	return &ImportHandler{workflowRepo: workflowRepo}
}

func (h *ImportHandler) Handle(w http.ResponseWriter, r *http.Request) {
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

	var imported ExportedWorkflow
	if err := json.NewDecoder(r.Body).Decode(&imported); err != nil {
		common.BadRequest(w, "Invalid workflow JSON")
		return
	}

	if errors := validation.Validate(imported); len(errors) > 0 {
		details := make([]common.ValidationDetail, len(errors))
		for i, e := range errors {
			details[i] = common.ValidationDetail{Field: e.Field, Message: e.Message}
		}
		common.ValidationErrors(w, details)
		return
	}

	wf, err := workflow.NewWorkflow(wsCtx.WorkspaceID, userClaims.UserID, imported.Name)
	if err != nil {
		common.HandleError(w, err)
		return
	}
	wf.Description = imported.Description
	wf.Nodes = imported.Nodes
	wf.Connections = imported.Connections
	wf.Settings = imported.Settings
	wf.Tags = imported.Tags

	if err := h.workflowRepo.Create(r.Context(), wf); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Created(w, ToWorkflowResponse(wf))
}
