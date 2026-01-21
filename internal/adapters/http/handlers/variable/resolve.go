package variable

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	domainvar "github.com/linkflow-ai/linkflow/internal/core/domain/variable"
)

// ResolveHandler handles resolving variables for an environment
type ResolveHandler struct {
	repo domainvar.Repository
}

// NewResolveHandler creates a new resolve handler
func NewResolveHandler(repo domainvar.Repository) *ResolveHandler {
	return &ResolveHandler{repo: repo}
}

// Handle handles the resolve variables request
func (h *ResolveHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceID := middleware.GetWorkspaceID(ctx)

	env := r.URL.Query().Get("environment")
	if env == "" {
		env = "development"
	}

	var workflowID *uuid.UUID
	if wfID := r.URL.Query().Get("workflow_id"); wfID != "" {
		if id, err := uuid.Parse(wfID); err == nil {
			workflowID = &id
		}
	}

	varSet, err := h.repo.ResolveVariables(ctx, workspaceID, env, workflowID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, ResolveResponse{
		Variables:   varSet.Variables,
		Environment: varSet.Environment,
	})
}
