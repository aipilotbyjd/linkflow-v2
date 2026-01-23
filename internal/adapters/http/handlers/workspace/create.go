package workspace

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	workspaceCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/workspace"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workspace"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/validation"
)

// CreateRequest represents workspace creation request
type CreateRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	Logo        *string `json:"logo,omitempty" validate:"omitempty,url,max=500"`
}

// WorkspaceResponse represents workspace in responses
type WorkspaceResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
	Logo        *string `json:"logo,omitempty"`
	OwnerID     string  `json:"owner_id"`
	Plan        string  `json:"plan"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// CreateHandler handles workspace creation
type CreateHandler struct {
	handler *workspaceCmd.CreateWorkspaceHandler
}

// NewCreateHandler creates a new handler
func NewCreateHandler(handler *workspaceCmd.CreateWorkspaceHandler) *CreateHandler {
	return &CreateHandler{handler: handler}
}

// Handle handles the create workspace request
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

	userClaims := middleware.GetUserFromContext(r.Context())
	if userClaims == nil {
		common.Unauthorized(w, "")
		return
	}

	ws, err := h.handler.Handle(r.Context(), workspaceCmd.CreateWorkspaceCommand{
		OwnerID:     userClaims.UserID,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Created(w, toWorkspaceResponse(ws))
}

func toWorkspaceResponse(ws *workspace.Workspace) WorkspaceResponse {
	return WorkspaceResponse{
		ID:          ws.ID.String(),
		Name:        ws.Name,
		Slug:        ws.Slug,
		Description: ws.Description,
		Logo:        ws.LogoURL,
		OwnerID:     ws.OwnerID.String(),
		Plan:        ws.PlanID,
		CreatedAt:   ws.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   ws.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
