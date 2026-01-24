package invitation

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workspace"
)

type GetInvitationHandler struct {
	invitationRepo workspace.InvitationRepository
}

func NewGetInvitationHandler(invitationRepo workspace.InvitationRepository) *GetInvitationHandler {
	return &GetInvitationHandler{invitationRepo: invitationRepo}
}

func (h *GetInvitationHandler) Handle(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		common.BadRequest(w, "Token is required")
		return
	}

	invitation, err := h.invitationRepo.FindByToken(r.Context(), token)
	if err != nil {
		common.HandleError(w, err)
		return
	}
	if invitation == nil {
		common.NotFound(w, "Invitation not found")
		return
	}

	// Return public info about the invitation (e.g. Workspace Name, Email)
	// We don't have Workspace Name in Invitation model directly, might need to preload or fetch.
	// For now return basic info.

	common.Success(w, map[string]interface{}{
		"email":        invitation.Email,
		"workspace_id": invitation.WorkspaceID,
		"valid":        invitation.IsValid(),
	})
}
