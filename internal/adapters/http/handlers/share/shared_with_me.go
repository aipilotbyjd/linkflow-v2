package share

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/share"
)

// SharedWithMeHandler handles get shares shared with user request
type SharedWithMeHandler struct {
	repo share.Repository
}

// NewSharedWithMeHandler creates a new handler
func NewSharedWithMeHandler(repo share.Repository) *SharedWithMeHandler {
	return &SharedWithMeHandler{repo: repo}
}

// Handle handles the get shares shared with user request
func (h *SharedWithMeHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	shares, err := h.repo.FindSharedWithUser(r.Context(), userID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, map[string]interface{}{
		"shares": ToShareResponseList(shares),
	})
}
