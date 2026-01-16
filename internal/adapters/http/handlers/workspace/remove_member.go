package workspace

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

type RemoveMemberHandler struct{}

func NewRemoveMemberHandler() *RemoveMemberHandler {
	return &RemoveMemberHandler{}
}

func (h *RemoveMemberHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "id")
	memberID := chi.URLParam(r, "memberId")

	if workspaceID == "" || memberID == "" {
		common.BadRequest(w, "Workspace ID and Member ID are required")
		return
	}

	// TODO: Implement member removal
	// 1. Verify current user has permission
	// 2. Cannot remove owner
	// 3. Remove member from workspace

	common.Success(w, map[string]string{
		"message": "Member removed successfully",
	})
}
