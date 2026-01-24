package workspace

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/rbac"
	"github.com/linkflow-ai/linkflow/internal/core/domain/user"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workspace"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/email"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/validation"
)

type InviteMemberRequest struct {
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"required,workspace_role"`
}

type InviteMemberHandler struct {
	workspaceRepo  workspace.Repository
	memberRepo     workspace.MemberRepository
	invitationRepo workspace.InvitationRepository
	userRepo       user.Repository
	rbacRepo       rbac.Repository
	emailService   email.Provider
	baseURL        string
}

func NewInviteMemberHandler(
	workspaceRepo workspace.Repository,
	memberRepo workspace.MemberRepository,
	invitationRepo workspace.InvitationRepository,
	userRepo user.Repository,
	rbacRepo rbac.Repository,
	emailService email.Provider,
	baseURL string,
) *InviteMemberHandler {
	return &InviteMemberHandler{
		workspaceRepo:  workspaceRepo,
		memberRepo:     memberRepo,
		invitationRepo: invitationRepo,
		userRepo:       userRepo,
		rbacRepo:       rbacRepo,
		emailService:   emailService,
		baseURL:        baseURL,
	}
}

func (h *InviteMemberHandler) Handle(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		common.BadRequest(w, "workspace context required")
		return
	}
	workspaceID := wsCtx.WorkspaceID

	var req InviteMemberRequest
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

	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		common.Unauthorized(w, "authentication required")
		return
	}

	// Verify current user has permission to invite
	inviter, err := h.memberRepo.FindByWorkspaceAndUser(r.Context(), workspaceID, claims.UserID)
	if err != nil {
		common.Forbidden(w, "You are not a member of this workspace")
		return
	}

	if !inviter.Role.CanInviteMembers() {
		common.Forbidden(w, "You don't have permission to invite members")
		return
	}

	// Get workspace info
	ws, err := h.workspaceRepo.FindByID(r.Context(), workspaceID)
	if err != nil {
		common.NotFound(w, "Workspace not found")
		return
	}

	// Check if user exists
	existingUser, err := h.userRepo.FindByEmail(r.Context(), req.Email)
	if err == nil && existingUser != nil {
		// Check if already a member
		_, err := h.memberRepo.FindByWorkspaceAndUser(r.Context(), workspaceID, existingUser.ID)
		if err == nil {
			common.BadRequest(w, "User is already a member of this workspace")
			return
		}

		// Add member directly
		role := workspace.Role(req.Role)
		member, err := workspace.NewMember(workspaceID, existingUser.ID, role)
		if err != nil {
			common.HandleError(w, err)
			return
		}

		// Set RoleID based on legacy role string
		// Map: "admin" -> "Admin", "member" -> "Editor", "viewer" -> "Viewer"
		var rbacRoleName string
		switch role {
		case workspace.RoleAdmin:
			rbacRoleName = rbac.RoleAdmin
		case workspace.RoleMember:
			rbacRoleName = rbac.RoleEditor // Map member to Editor in RBAC
		case workspace.RoleViewer:
			rbacRoleName = rbac.RoleViewer
		default:
			rbacRoleName = rbac.RoleViewer
		}

		if rbacRole, err := h.rbacRepo.GetRoleByName(r.Context(), nil, rbacRoleName); err == nil {
			member.RoleID = &rbacRole.ID
		}

		if err := h.memberRepo.Create(r.Context(), member); err != nil {
			common.HandleError(w, err)
			return
		}

		common.Success(w, map[string]string{
			"message": "Member added successfully",
		})
		return
	}

	// User doesn't exist - send invitation email
	// Generate token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		common.HandleError(w, err)
		return
	}
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	// Create invitation
	invitation, err := workspace.NewInvitation(workspaceID, req.Email, workspace.Role(req.Role), inviter.UserID, token, 7*24*time.Hour)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	// Set RoleID for invitation
	// Map legacy role string to RBAC Role
	var rbacRoleName string
	role := workspace.Role(req.Role)
	switch role {
	case workspace.RoleAdmin:
		rbacRoleName = rbac.RoleAdmin
	case workspace.RoleMember:
		rbacRoleName = rbac.RoleEditor
	case workspace.RoleViewer:
		rbacRoleName = rbac.RoleViewer
	default:
		rbacRoleName = rbac.RoleViewer
	}

	if rbacRole, err := h.rbacRepo.GetRoleByName(r.Context(), nil, rbacRoleName); err == nil {
		invitation.RoleID = &rbacRole.ID
	}

	// Save invitation
	if err := h.invitationRepo.Create(r.Context(), invitation); err != nil {
		common.HandleError(w, err)
		return
	}

	inviteURL := h.baseURL + "/invite?token=" + token
	msg := &email.Message{
		To:      []string{req.Email},
		Subject: "You've been invited to join " + ws.Name,
		HTMLBody: `<p>Hello,</p>
<p>You've been invited to join the workspace "` + ws.Name + `" on LinkFlow.</p>
<p><a href="` + inviteURL + `">Accept Invitation</a></p>
<p>If you don't have an account yet, you'll need to create one first.</p>`,
		TextBody: "Hello,\n\nYou've been invited to join the workspace \"" + ws.Name + "\" on LinkFlow.\n\nAccept your invitation here: " + inviteURL,
	}
	_ = h.emailService.Send(r.Context(), msg)

	common.Success(w, map[string]string{
		"message": "Invitation sent successfully",
	})
}
