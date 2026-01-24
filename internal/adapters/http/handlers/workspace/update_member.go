package workspace

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/rbac"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workspace"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/validation"
)

type UpdateMemberRequest struct {
	Role string `json:"role" validate:"required,oneof=owner admin editor viewer"`
}

type UpdateMemberHandler struct {
	memberRepo    workspace.MemberRepository
	workspaceRepo workspace.Repository
	rbacRepo      rbac.Repository
}

func NewUpdateMemberHandler(
	memberRepo workspace.MemberRepository,
	workspaceRepo workspace.Repository,
	rbacRepo rbac.Repository,
) *UpdateMemberHandler {
	return &UpdateMemberHandler{
		memberRepo:    memberRepo,
		workspaceRepo: workspaceRepo,
		rbacRepo:      rbacRepo,
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
		common.BadRequest(w, "Invalid request body")
		return
	}

	if errors := validation.Validate(req); len(errors) > 0 {
		details := make([]common.ValidationDetail, len(errors))
		for i, e := range errors {
			details[i] = common.ValidationDetail{Field: e.Field, Message: e.Message}
		}
		common.ValidationErrors(w, details)
		return
	}

	role, err := workspace.ParseRole(req.Role)
	if err != nil {
		common.BadRequest(w, "Invalid role: must be one of owner, admin, editor, viewer")
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

	// Update RoleID based on legacy role string
	var rbacRoleName string
	switch role {
	case workspace.RoleAdmin:
		rbacRoleName = rbac.RoleAdmin
	case workspace.RoleMember:
		rbacRoleName = rbac.RoleEditor
	case workspace.RoleViewer:
		rbacRoleName = rbac.RoleViewer
	case workspace.RoleOwner:
		rbacRoleName = rbac.RoleOwner
	default:
		rbacRoleName = rbac.RoleViewer
	}

	if rbacRole, err := h.rbacRepo.GetRoleByName(r.Context(), nil, rbacRoleName); err == nil {
		member.RoleID = &rbacRole.ID
	}

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
