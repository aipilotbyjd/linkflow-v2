package share

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

// UpdateShareRequest represents update share request
type UpdateShareRequest struct {
	Permission string `json:"permission"`
}

// UpdateHandler handles update share request
type UpdateHandler struct {
	repo ShareRepository
}

// NewUpdateHandler creates a new handler
func NewUpdateHandler(repo ShareRepository) *UpdateHandler {
	return &UpdateHandler{repo: repo}
}

// Handle handles the update share request
func (h *UpdateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	shareID := chi.URLParam(r, "shareId")

	var req UpdateShareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "invalid request body")
		return
	}

	share := WorkflowShare{
		ID:         shareID,
		Permission: req.Permission,
		Status:     "accepted",
	}

	common.Success(w, share)
}
