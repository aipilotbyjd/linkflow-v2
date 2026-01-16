package workspace

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workspace"
)

type InviteMemberCommand struct {
	WorkspaceID uuid.UUID
	InviterID   uuid.UUID
	Email       string
	Role        string
}

type InviteMemberHandler struct {
	workspaceRepo workspace.Repository
	memberRepo    workspace.MemberRepository
}

func NewInviteMemberHandler(
	workspaceRepo workspace.Repository,
	memberRepo workspace.MemberRepository,
) *InviteMemberHandler {
	return &InviteMemberHandler{
		workspaceRepo: workspaceRepo,
		memberRepo:    memberRepo,
	}
}

func (h *InviteMemberHandler) Handle(ctx context.Context, cmd InviteMemberCommand) error {
	// Check inviter permission
	inviter, err := h.memberRepo.FindByWorkspaceAndUser(ctx, cmd.WorkspaceID, cmd.InviterID)
	if err != nil {
		return workspace.ErrNotAMember
	}

	if !inviter.Role.CanInviteMembers() {
		return workspace.ErrPermissionDenied
	}

	// TODO: Create invitation record
	// TODO: Send invitation email

	return nil
}
