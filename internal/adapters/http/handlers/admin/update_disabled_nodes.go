package admin

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/sitesettings"
)

// UpdateDisabledNodesRequest represents the request body
type UpdateDisabledNodesRequest struct {
	DisabledNodes []string `json:"disabled_nodes"`
}

// UpdateDisabledNodesHandler handles update disabled nodes request
type UpdateDisabledNodesHandler struct {
	repo sitesettings.Repository
}

// NewUpdateDisabledNodesHandler creates a new handler
func NewUpdateDisabledNodesHandler(repo sitesettings.Repository) *UpdateDisabledNodesHandler {
	return &UpdateDisabledNodesHandler{repo: repo}
}

// Handle handles the request
func (h *UpdateDisabledNodesHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req UpdateDisabledNodesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}

	// Ensure we have a valid slice (not nil)
	if req.DisabledNodes == nil {
		req.DisabledNodes = []string{}
	}

	if err := h.repo.SetDisabledNodes(r.Context(), req.DisabledNodes); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, map[string]interface{}{
		"message":        "Disabled nodes updated successfully",
		"disabled_nodes": req.DisabledNodes,
	})
}
