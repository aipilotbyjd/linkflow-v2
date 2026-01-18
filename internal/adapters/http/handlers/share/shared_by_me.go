package share

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/share"
)

// SharedByMeHandler handles get shares created by user request
type SharedByMeHandler struct {
	repo share.Repository
}

// NewSharedByMeHandler creates a new handler
func NewSharedByMeHandler(repo share.Repository) *SharedByMeHandler {
	return &SharedByMeHandler{repo: repo}
}

// Handle handles the get shares created by user request
func (h *SharedByMeHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	shares, err := h.repo.FindSharedByUser(r.Context(), userID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, map[string]interface{}{
		"shares": ToShareResponseList(shares),
	})
}
