package webhook

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

type ActivateEndpointHandler struct{}

func NewActivateEndpointHandler() *ActivateEndpointHandler {
	return &ActivateEndpointHandler{}
}

func (h *ActivateEndpointHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "id")
	endpointID := chi.URLParam(r, "endpointId")

	if workspaceID == "" || endpointID == "" {
		common.BadRequest(w, "Workspace ID and Endpoint ID are required")
		return
	}

	// TODO: Implement endpoint activation
	// 1. Verify endpoint exists and belongs to workspace
	// 2. Set is_active = true

	common.Success(w, map[string]string{
		"message": "Webhook endpoint activated",
	})
}
