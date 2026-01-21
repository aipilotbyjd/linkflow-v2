package variable

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	domainvar "github.com/linkflow-ai/linkflow/internal/core/domain/variable"
)

// UpdateHandler handles updating variables
type UpdateHandler struct {
	repo domainvar.Repository
}

// NewUpdateHandler creates a new update handler
func NewUpdateHandler(repo domainvar.Repository) *UpdateHandler {
	return &UpdateHandler{repo: repo}
}

// Handle handles the update variable request
func (h *UpdateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	varID, err := uuid.Parse(chi.URLParam(r, "variableId"))
	if err != nil {
		common.BadRequest(w, "Invalid variable ID")
		return
	}

	var req VariableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}

	v, err := h.repo.FindVariableByID(ctx, varID)
	if err != nil {
		common.NotFound(w, "variable")
		return
	}

	v.Update(req.Value, req.Description)
	if req.IsSecret && !v.IsSecret {
		v.MarkAsSecret()
	}

	if err := h.repo.UpdateVariable(ctx, v); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, ToVariableResponse(v))
}
