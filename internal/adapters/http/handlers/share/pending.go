package share

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/share"
)

// PendingHandler handles get pending shares request
type PendingHandler struct {
	repo share.Repository
}

// NewPendingHandler creates a new handler
func NewPendingHandler(repo share.Repository) *PendingHandler {
	return &PendingHandler{repo: repo}
}

// Handle handles the get pending shares request
func (h *PendingHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	shares, err := h.repo.FindPendingForUser(r.Context(), userID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, map[string]interface{}{
		"shares": ToShareResponseList(shares),
	})
}
