package variable

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	domainvar "github.com/linkflow-ai/linkflow/internal/core/domain/variable"
)

// CreateHandler handles creating variables
type CreateHandler struct {
	repo domainvar.Repository
}

// NewCreateHandler creates a new create handler
func NewCreateHandler(repo domainvar.Repository) *CreateHandler {
	return &CreateHandler{repo: repo}
}

// Handle handles the create variable request
func (h *CreateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceID := middleware.GetWorkspaceID(ctx)
	userID := middleware.GetUserID(ctx)

	var req VariableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}

	if req.Key == "" {
		common.BadRequest(w, "Key is required")
		return
	}

	// Check if key already exists
	existing, _ := h.repo.FindVariableByKey(ctx, workspaceID, req.Key)
	if existing != nil {
		common.BadRequest(w, "Variable key already exists")
		return
	}

	v := domainvar.NewVariable(workspaceID, userID, req.Key, req.Value)
	v.Description = req.Description
	if req.IsSecret {
		v.MarkAsSecret()
	}

	if err := h.repo.CreateVariable(ctx, v); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Created(w, ToVariableResponse(v))
}
