package workspace

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workspace"
)

type UpdateMemberRequest struct {
	Role string `json:"role"`
}

type UpdateMemberHandler struct {
	memberRepo    workspace.MemberRepository
	workspaceRepo workspace.Repository
}

func NewUpdateMemberHandler(memberRepo workspace.MemberRepository, workspaceRepo workspace.Repository) *UpdateMemberHandler {
	return &UpdateMemberHandler{
		memberRepo:    memberRepo,
		workspaceRepo: workspaceRepo,
	}
}

func (h *UpdateMemberHandler) Handle(w http.ResponseWriter, r *http.Request) {
	memberIDStr := chi.URLParam(r, "memberId")
	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		common.BadRequest(w, "invalid member ID")
		return
	}

	workspaceID := middleware.GetWorkspaceID(r.Context())
	currentUserID := middleware.GetUserID(r.Context())

	var req UpdateMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "invalid request body")
		return
	}

	role, err := workspace.ParseRole(req.Role)
	if err != nil {
		common.BadRequest(w, "invalid role: must be one of owner, admin, editor, viewer")
		return
	}

	currentMember, err := h.memberRepo.FindByWorkspaceAndUser(r.Context(), workspaceID, currentUserID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	if currentMember.Role != workspace.RoleOwner && currentMember.Role != workspace.RoleAdmin {
		common.Forbidden(w, "only owners and admins can update member roles")
		return
	}

	member, err := h.memberRepo.FindByID(r.Context(), memberID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	if member.WorkspaceID.String() != workspaceID.String() {
		common.NotFound(w, "member not found")
		return
	}

	if member.Role == workspace.RoleOwner && role != workspace.RoleOwner {
		owners, _, err := h.memberRepo.FindByWorkspaceID(r.Context(), workspaceID, nil)
		if err != nil {
			common.HandleError(w, err)
			return
		}
		
		ownerCount := 0
		for _, m := range owners {
			if m.Role == workspace.RoleOwner {
				ownerCount++
			}
		}
		
		if ownerCount <= 1 {
			common.Error(w, http.StatusBadRequest, "LAST_OWNER", "Cannot demote the last owner")
			return
		}
	}

	if currentMember.Role != workspace.RoleOwner && role == workspace.RoleOwner {
		common.Forbidden(w, "only owners can promote members to owner")
		return
	}

	member.Role = role
	if err := h.memberRepo.Update(r.Context(), member); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, map[string]interface{}{
		"id":          member.ID.String(),
		"userId":      member.UserID.String(),
		"workspaceId": member.WorkspaceID.String(),
		"role":        member.Role.String(),
	})
}
