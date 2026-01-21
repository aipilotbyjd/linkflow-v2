package variable

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	domainvar "github.com/linkflow-ai/linkflow/internal/core/domain/variable"
)

// ListEnvironmentsHandler handles listing environments
type ListEnvironmentsHandler struct {
	repo domainvar.Repository
}

// NewListEnvironmentsHandler creates a new list environments handler
func NewListEnvironmentsHandler(repo domainvar.Repository) *ListEnvironmentsHandler {
	return &ListEnvironmentsHandler{repo: repo}
}

// Handle handles the list environments request
func (h *ListEnvironmentsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceID := middleware.GetWorkspaceID(ctx)

	environments, err := h.repo.FindEnvironmentsByWorkspace(ctx, workspaceID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	responses := make([]EnvironmentResponse, len(environments))
	for i, e := range environments {
		responses[i] = ToEnvironmentResponse(&e)
	}

	common.Success(w, map[string]interface{}{
		"environments": responses,
		"total":        len(responses),
	})
}
