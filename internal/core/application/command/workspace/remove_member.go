package workspace

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workspace"
)

type RemoveMemberCommand struct {
	WorkspaceID uuid.UUID
	RemoverID   uuid.UUID
	MemberID    uuid.UUID
}

type RemoveMemberHandler struct {
	workspaceRepo workspace.Repository
	memberRepo    workspace.MemberRepository
}

func NewRemoveMemberHandler(
	workspaceRepo workspace.Repository,
	memberRepo workspace.MemberRepository,
) *RemoveMemberHandler {
	return &RemoveMemberHandler{
		workspaceRepo: workspaceRepo,
		memberRepo:    memberRepo,
	}
}

func (h *RemoveMemberHandler) Handle(ctx context.Context, cmd RemoveMemberCommand) error {
	ws, err := h.workspaceRepo.FindByID(ctx, cmd.WorkspaceID)
	if err != nil {
		return err
	}

	// Cannot remove owner
	if ws.OwnerID == cmd.MemberID {
		return workspace.ErrCannotRemoveOwner
	}

	// Check remover permission
	remover, err := h.memberRepo.FindByWorkspaceAndUser(ctx, cmd.WorkspaceID, cmd.RemoverID)
	if err != nil {
		return workspace.ErrNotAMember
	}

	if !remover.Role.CanRemoveMembers() {
		return workspace.ErrPermissionDenied
	}

	return h.memberRepo.DeleteByWorkspaceAndUser(ctx, cmd.WorkspaceID, cmd.MemberID)
}
