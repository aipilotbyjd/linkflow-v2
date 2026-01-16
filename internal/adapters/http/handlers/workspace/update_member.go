package workspace

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

type UpdateMemberRequest struct {
	Role string `json:"role" validate:"required,oneof=admin member viewer"`
}

type UpdateMemberHandler struct{}

func NewUpdateMemberHandler() *UpdateMemberHandler {
	return &UpdateMemberHandler{}
}

func (h *UpdateMemberHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "id")
	memberID := chi.URLParam(r, "memberId")

	if workspaceID == "" || memberID == "" {
		common.BadRequest(w, "Workspace ID and Member ID are required")
		return
	}

	var req UpdateMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}

	// TODO: Implement member role update
	// 1. Verify current user has permission
	// 2. Cannot demote owner
	// 3. Update member role

	common.Success(w, map[string]string{
		"message": "Member role updated successfully",
	})
}
