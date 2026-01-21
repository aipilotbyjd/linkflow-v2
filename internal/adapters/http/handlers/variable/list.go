package variable

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	domainvar "github.com/linkflow-ai/linkflow/internal/core/domain/variable"
)

// ListHandler handles listing variables
type ListHandler struct {
	repo domainvar.Repository
}

// NewListHandler creates a new list handler
func NewListHandler(repo domainvar.Repository) *ListHandler {
	return &ListHandler{repo: repo}
}

// Handle handles the list variables request
func (h *ListHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceID := middleware.GetWorkspaceID(ctx)

	variables, err := h.repo.FindVariablesByWorkspace(ctx, workspaceID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	responses := make([]VariableResponse, len(variables))
	for i, v := range variables {
		responses[i] = ToVariableResponse(&v)
	}

	common.Success(w, map[string]interface{}{
		"variables": responses,
		"total":     len(responses),
	})
}
