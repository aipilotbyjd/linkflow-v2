package webhook

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/webhook"
)

type RegenerateSecretHandler struct {
	webhookRepo webhook.Repository
}

func NewRegenerateSecretHandler(webhookRepo webhook.Repository) *RegenerateSecretHandler {
	return &RegenerateSecretHandler{webhookRepo: webhookRepo}
}

func (h *RegenerateSecretHandler) Handle(w http.ResponseWriter, r *http.Request) {
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

	// Generate new secret
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		common.HandleError(w, err)
		return
	}
	newSecret := base64.URLEncoding.EncodeToString(secretBytes)

	// Update endpoint with new secret
	endpoint.RegenerateSecret(newSecret)
	if err := h.webhookRepo.Update(r.Context(), endpoint); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, map[string]string{
		"secret": newSecret,
	})
}
