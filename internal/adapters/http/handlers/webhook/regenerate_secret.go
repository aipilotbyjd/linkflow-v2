package webhook

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

type RegenerateSecretHandler struct{}

func NewRegenerateSecretHandler() *RegenerateSecretHandler {
	return &RegenerateSecretHandler{}
}

func (h *RegenerateSecretHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "id")
	endpointID := chi.URLParam(r, "endpointId")

	if workspaceID == "" || endpointID == "" {
		common.BadRequest(w, "Workspace ID and Endpoint ID are required")
		return
	}

	// TODO: Implement secret regeneration
	// 1. Verify endpoint exists and belongs to workspace
	// 2. Generate new secret
	// 3. Update endpoint

	common.Success(w, map[string]string{
		"secret": "new-generated-secret",
	})
}
