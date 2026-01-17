package webhook

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/webhook"
)

type DeactivateEndpointHandler struct {
	webhookRepo webhook.Repository
}

func NewDeactivateEndpointHandler(webhookRepo webhook.Repository) *DeactivateEndpointHandler {
	return &DeactivateEndpointHandler{webhookRepo: webhookRepo}
}

func (h *DeactivateEndpointHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workspaceIDStr := chi.URLParam(r, "id")
	endpointIDStr := chi.URLParam(r, "endpointId")

	if workspaceIDStr == "" || endpointIDStr == "" {
		common.BadRequest(w, "Workspace ID and Endpoint ID are required")
		return
	}

	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		common.BadRequest(w, "Invalid workspace ID")
		return
	}

	endpointID, err := uuid.Parse(endpointIDStr)
	if err != nil {
		common.BadRequest(w, "Invalid endpoint ID")
		return
	}

	// Verify endpoint exists and belongs to workspace
	endpoint, err := h.webhookRepo.FindByID(r.Context(), endpointID)
	if err != nil {
		common.NotFound(w, "Webhook endpoint not found")
		return
	}

	if endpoint.WorkspaceID != workspaceID {
		common.Forbidden(w, "Endpoint does not belong to this workspace")
		return
	}

	// Deactivate endpoint
	if err := h.webhookRepo.SetActive(r.Context(), endpointID, false); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, map[string]string{
		"message": "Webhook endpoint deactivated",
	})
}
