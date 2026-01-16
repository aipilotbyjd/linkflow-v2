package workspace

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	workspaceQuery "github.com/linkflow-ai/linkflow/internal/core/application/query/workspace"
)

// ListHandler handles listing workspaces
type ListHandler struct {
	handler *workspaceQuery.ListWorkspacesHandler
}

// NewListHandler creates a new handler
func NewListHandler(handler *workspaceQuery.ListWorkspacesHandler) *ListHandler {
	return &ListHandler{handler: handler}
}

// Handle handles the list workspaces request
func (h *ListHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userClaims := middleware.GetUserFromContext(r.Context())
	if userClaims == nil {
		common.Unauthorized(w, "")
		return
	}

	workspaces, err := h.handler.Handle(r.Context(), workspaceQuery.ListWorkspacesQuery{
		UserID: userClaims.UserID,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	responses := make([]WorkspaceResponse, len(workspaces))
	for i, ws := range workspaces {
		responses[i] = toWorkspaceResponse(&ws)
	}

	common.Success(w, responses)
}
