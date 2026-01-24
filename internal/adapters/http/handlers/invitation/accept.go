package invitation

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/user"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workspace"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/validation"
)

type AcceptInvitationRequest struct {
	Token string `json:"token" validate:"required"`
}

type AcceptInvitationHandler struct {
	invitationRepo workspace.InvitationRepository
	memberRepo     workspace.MemberRepository
	userRepo       user.Repository
}

func NewAcceptInvitationHandler(
	invitationRepo workspace.InvitationRepository,
	memberRepo workspace.MemberRepository,
	userRepo user.Repository,
) *AcceptInvitationHandler {
	return &AcceptInvitationHandler{
		invitationRepo: invitationRepo,
		memberRepo:     memberRepo,
		userRepo:       userRepo,
	}
}

func (h *AcceptInvitationHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req AcceptInvitationRequest
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

	userID := middleware.GetUserID(r.Context())
	if userID.String() == "" {
		common.Unauthorized(w, "Authentication required to accept invitation")
		return
	}

	// 1. Find Invitation
	invitation, err := h.invitationRepo.FindByToken(r.Context(), req.Token)
	if err != nil {
		common.HandleError(w, err)
		return
	}
	if invitation == nil {
		common.NotFound(w, "Invitation not found")
		return
	}

	// 2. Validate Invitation
	if !invitation.IsValid() {
		common.BadRequest(w, "Invitation is invalid or expired")
		return
	}

	// 3. Verify Email match
	currentUser, err := h.userRepo.FindByID(r.Context(), userID)
	if err != nil {
		common.HandleError(w, err)
		return
	}
	if currentUser.Email != invitation.Email {
		common.Forbidden(w, "Invitation email does not match logged in user")
		return
	}

	// 4. Check if already member
	isMember, err := h.memberRepo.IsMember(r.Context(), invitation.WorkspaceID, userID)
	if err != nil {
		common.HandleError(w, err)
		return
	}
	if isMember {
		// Already member - just cleanup invitation
		_ = h.invitationRepo.Delete(r.Context(), invitation.ID)
		common.Success(w, map[string]string{"message": "You are already a member of this workspace"})
		return
	}

	// 5. Create Member
	member, err := workspace.NewMember(invitation.WorkspaceID, userID, invitation.Role)
	if err != nil {
		common.HandleError(w, err)
		return
	}
	member.RoleID = invitation.RoleID // Assign custom role if present
	member.InvitedBy = &invitation.InvitedBy
	now := time.Now()
	member.JoinedAt = &now

	if err := h.memberRepo.Create(r.Context(), member); err != nil {
		common.HandleError(w, err)
		return
	}

	// 6. Mark Accepted / Delete Invitation
	// Usually better to delete one-time use tokens
	if err := h.invitationRepo.Delete(r.Context(), invitation.ID); err != nil {
		// Log warning, but don't fail request
	}

	common.Success(w, map[string]string{
		"message":      "Invitation accepted",
		"workspace_id": invitation.WorkspaceID.String(),
	})
}
