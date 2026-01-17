package workspace

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workspace"
)

type RemoveMemberHandler struct {
	workspaceRepo workspace.Repository
	memberRepo    workspace.MemberRepository
}

func NewRemoveMemberHandler(workspaceRepo workspace.Repository, memberRepo workspace.MemberRepository) *RemoveMemberHandler {
	return &RemoveMemberHandler{
		workspaceRepo: workspaceRepo,
		memberRepo:    memberRepo,
	}
}

func (h *RemoveMemberHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workspaceIDStr := chi.URLParam(r, "id")
	memberIDStr := chi.URLParam(r, "memberId")

	if workspaceIDStr == "" || memberIDStr == "" {
		common.BadRequest(w, "Workspace ID and Member ID are required")
		return
	}

	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		common.BadRequest(w, "Invalid workspace ID")
		return
	}

	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		common.BadRequest(w, "Invalid member ID")
		return
	}

	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		common.Unauthorized(w, "authentication required")
		return
	}

	// Verify current user has permission
	currentMember, err := h.memberRepo.FindByWorkspaceAndUser(r.Context(), workspaceID, claims.UserID)
	if err != nil {
		common.Forbidden(w, "You are not a member of this workspace")
		return
	}

	if !currentMember.Role.CanRemoveMembers() {
		common.Forbidden(w, "You don't have permission to remove members")
		return
	}

	// Get the member to be removed
	memberToRemove, err := h.memberRepo.FindByID(r.Context(), memberID)
	if err != nil {
		common.NotFound(w, "Member not found")
		return
	}

	if memberToRemove.WorkspaceID != workspaceID {
		common.Forbidden(w, "Member does not belong to this workspace")
		return
	}

	// Cannot remove owner
	if memberToRemove.Role == workspace.RoleOwner {
		common.BadRequest(w, "Cannot remove the workspace owner")
		return
	}

	// Cannot remove yourself (use leave workspace instead)
	if memberToRemove.UserID == claims.UserID {
		common.BadRequest(w, "Cannot remove yourself. Use leave workspace instead.")
		return
	}

	// Remove member
	if err := h.memberRepo.Delete(r.Context(), memberID); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, map[string]string{
		"message": "Member removed successfully",
	})
}
