package workspace

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	workspaceQuery "github.com/linkflow-ai/linkflow/internal/core/application/query/workspace"
)

// GetHandler handles getting a single workspace
type GetHandler struct {
	handler *workspaceQuery.GetWorkspaceHandler
}

// NewGetHandler creates a new handler
func NewGetHandler(handler *workspaceQuery.GetWorkspaceHandler) *GetHandler {
	return &GetHandler{handler: handler}
}

// Handle handles the get workspace request
func (h *GetHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workspaceIDStr := chi.URLParam(r, "workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		common.BadRequest(w, "invalid workspace ID")
		return
	}

	ws, err := h.handler.Handle(r.Context(), workspaceQuery.GetWorkspaceQuery{
		WorkspaceID: workspaceID,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, toWorkspaceResponse(ws))
}
