package byok

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

// DeleteHandler handles deleting BYOK configurations
type DeleteHandler struct {
	repo BYOKRepository
}

func NewDeleteHandler(repo BYOKRepository) *DeleteHandler {
	return &DeleteHandler{repo: repo}
}

func (h *DeleteHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	configID, err := uuid.Parse(chi.URLParam(r, "configId"))
	if err != nil {
		common.BadRequest(w, "Invalid configuration ID")
		return
	}

	if err := h.repo.Delete(ctx, configID); err != nil {
		common.HandleError(w, err)
		return
	}

	common.NoContent(w)
}
