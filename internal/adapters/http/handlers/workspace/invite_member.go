package workspace

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

type InviteMemberRequest struct {
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"required,oneof=admin member viewer"`
}

type InviteMemberHandler struct{}

func NewInviteMemberHandler() *InviteMemberHandler {
	return &InviteMemberHandler{}
}

func (h *InviteMemberHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "id")
	if workspaceID == "" {
		common.BadRequest(w, "Workspace ID is required")
		return
	}

	var req InviteMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}

	// TODO: Implement member invitation
	// 1. Verify current user has permission to invite
	// 2. Check if user already a member
	// 3. Create invitation or add member if user exists
	// 4. Send invitation email

	common.Success(w, map[string]string{
		"message": "Invitation sent successfully",
	})
}
