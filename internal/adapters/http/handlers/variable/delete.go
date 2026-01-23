package variable

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	domainvar "github.com/linkflow-ai/linkflow/internal/core/domain/variable"
)

// DeleteHandler handles deleting variables
type DeleteHandler struct {
	repo domainvar.Repository
}

// NewDeleteHandler creates a new delete handler
func NewDeleteHandler(repo domainvar.Repository) *DeleteHandler {
	return &DeleteHandler{repo: repo}
}

// Handle handles the delete variable request
func (h *DeleteHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	varID, err := uuid.Parse(chi.URLParam(r, "variableId"))
	if err != nil {
		common.BadRequest(w, "Invalid variable ID")
		return
	}

	if err := h.repo.DeleteVariable(ctx, varID); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, map[string]string{"message": "Variable deleted"})
}
