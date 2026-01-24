package admin

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/sitesettings"
)

// GetDisabledNodesHandler handles get disabled nodes request
type GetDisabledNodesHandler struct {
	repo sitesettings.Repository
}

// NewGetDisabledNodesHandler creates a new handler
func NewGetDisabledNodesHandler(repo sitesettings.Repository) *GetDisabledNodesHandler {
	return &GetDisabledNodesHandler{repo: repo}
}

// Handle handles the request
func (h *GetDisabledNodesHandler) Handle(w http.ResponseWriter, r *http.Request) {
	nodes, err := h.repo.GetDisabledNodes(r.Context())
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, map[string]interface{}{
		"disabled_nodes": nodes,
	})
}
