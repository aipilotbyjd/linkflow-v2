package workspace

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workspace"
)

type UpdateWorkspaceCommand struct {
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
	Name        *string
	Description *string
}

type UpdateWorkspaceHandler struct {
	workspaceRepo workspace.Repository
	memberRepo    workspace.MemberRepository
}

func NewUpdateWorkspaceHandler(
	workspaceRepo workspace.Repository,
	memberRepo workspace.MemberRepository,
) *UpdateWorkspaceHandler {
	return &UpdateWorkspaceHandler{
		workspaceRepo: workspaceRepo,
		memberRepo:    memberRepo,
	}
}

func (h *UpdateWorkspaceHandler) Handle(ctx context.Context, cmd UpdateWorkspaceCommand) (*workspace.Workspace, error) {
	// Check permission
	member, err := h.memberRepo.FindByWorkspaceAndUser(ctx, cmd.WorkspaceID, cmd.UserID)
	if err != nil {
		return nil, workspace.ErrNotAMember
	}

	if !member.Role.CanManageWorkspace() {
		return nil, workspace.ErrPermissionDenied
	}

	ws, err := h.workspaceRepo.FindByID(ctx, cmd.WorkspaceID)
	if err != nil {
		return nil, err
	}

	if cmd.Name != nil {
		ws.Name = *cmd.Name
	}
	if cmd.Description != nil {
		ws.Description = cmd.Description
	}

	if err := h.workspaceRepo.Update(ctx, ws); err != nil {
		return nil, err
	}

	return ws, nil
}
