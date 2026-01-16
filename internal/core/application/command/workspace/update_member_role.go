package workspace

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workspace"
)

type UpdateMemberRoleCommand struct {
	WorkspaceID uuid.UUID
	UpdaterID   uuid.UUID
	MemberID    uuid.UUID
	NewRole     string
}

type UpdateMemberRoleHandler struct {
	workspaceRepo workspace.Repository
	memberRepo    workspace.MemberRepository
}

func NewUpdateMemberRoleHandler(
	workspaceRepo workspace.Repository,
	memberRepo workspace.MemberRepository,
) *UpdateMemberRoleHandler {
	return &UpdateMemberRoleHandler{
		workspaceRepo: workspaceRepo,
		memberRepo:    memberRepo,
	}
}

func (h *UpdateMemberRoleHandler) Handle(ctx context.Context, cmd UpdateMemberRoleCommand) error {
	ws, err := h.workspaceRepo.FindByID(ctx, cmd.WorkspaceID)
	if err != nil {
		return err
	}

	// Cannot change owner's role
	if ws.OwnerID == cmd.MemberID {
		return workspace.ErrCannotChangeOwnerRole
	}

	// Check updater permission
	updater, err := h.memberRepo.FindByWorkspaceAndUser(ctx, cmd.WorkspaceID, cmd.UpdaterID)
	if err != nil {
		return workspace.ErrNotAMember
	}

	if !updater.Role.CanManageRoles() {
		return workspace.ErrPermissionDenied
	}

	// Get member to update
	member, err := h.memberRepo.FindByWorkspaceAndUser(ctx, cmd.WorkspaceID, cmd.MemberID)
	if err != nil {
		return err
	}

	// Update role
	newRole, err := workspace.ParseRole(cmd.NewRole)
	if err != nil {
		return err
	}
	member.Role = newRole

	return h.memberRepo.Update(ctx, member)
}
