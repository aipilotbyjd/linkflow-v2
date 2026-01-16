package workflow

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	workflowCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/workflow"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// CreateWorkflowRequest represents workflow creation request
type CreateWorkflowRequest struct {
	Name        string          `json:"name" validate:"required"`
	Description *string         `json:"description,omitempty"`
	Nodes       types.JSONArray `json:"nodes"`
	Connections types.JSONArray `json:"connections"`
	Settings    types.JSON      `json:"settings"`
	Tags        []string        `json:"tags"`
	Color       *string         `json:"color,omitempty"`
	Icon        *string         `json:"icon,omitempty"`
	Category    *string         `json:"category,omitempty"`
}

// WorkflowResponse represents workflow in responses
type WorkflowResponse struct {
	ID             string          `json:"id"`
	WorkspaceID    string          `json:"workspace_id"`
	Name           string          `json:"name"`
	Description    *string         `json:"description,omitempty"`
	Status         string          `json:"status"`
	Version        int             `json:"version"`
	Nodes          types.JSONArray `json:"nodes"`
	Connections    types.JSONArray `json:"connections"`
	Settings       types.JSON      `json:"settings"`
	Tags           []string        `json:"tags"`
	Color          *string         `json:"color,omitempty"`
	Icon           *string         `json:"icon,omitempty"`
	Category       *string         `json:"category,omitempty"`
	IsFavorite     bool            `json:"is_favorite"`
	ExecutionCount int             `json:"execution_count"`
	CreatedAt      string          `json:"created_at"`
	UpdatedAt      string          `json:"updated_at"`
}

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
		common.BadRequest(w, "invalid request body")
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

	common.Created(w, toWorkflowResponse(wf))
}

func toWorkflowResponse(wf *workflow.Workflow) WorkflowResponse {
	return WorkflowResponse{
		ID:             wf.ID.String(),
		WorkspaceID:    wf.WorkspaceID.String(),
		Name:           wf.Name,
		Description:    wf.Description,
		Status:         string(wf.Status),
		Version:        wf.Version,
		Nodes:          wf.Nodes,
		Connections:    wf.Connections,
		Settings:       wf.Settings,
		Tags:           wf.Tags,
		Color:          wf.Color,
		Icon:           wf.Icon,
		Category:       wf.Category,
		IsFavorite:     wf.IsFavorite,
		ExecutionCount: wf.ExecutionCount,
		CreatedAt:      wf.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      wf.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
