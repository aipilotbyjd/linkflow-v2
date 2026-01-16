package webhook

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

type DeactivateEndpointHandler struct{}

func NewDeactivateEndpointHandler() *DeactivateEndpointHandler {
	return &DeactivateEndpointHandler{}
}

func (h *DeactivateEndpointHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "id")
	endpointID := chi.URLParam(r, "endpointId")

	if workspaceID == "" || endpointID == "" {
		common.BadRequest(w, "Workspace ID and Endpoint ID are required")
		return
	}

	// TODO: Implement endpoint deactivation
	// 1. Verify endpoint exists and belongs to workspace
	// 2. Set is_active = false

	common.Success(w, map[string]string{
		"message": "Webhook endpoint deactivated",
	})
}
