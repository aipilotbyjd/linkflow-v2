package workspace

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workspace"
)

type DeleteWorkspaceCommand struct {
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
}

type DeleteWorkspaceHandler struct {
	workspaceRepo workspace.Repository
}

func NewDeleteWorkspaceHandler(workspaceRepo workspace.Repository) *DeleteWorkspaceHandler {
	return &DeleteWorkspaceHandler{workspaceRepo: workspaceRepo}
}

func (h *DeleteWorkspaceHandler) Handle(ctx context.Context, cmd DeleteWorkspaceCommand) error {
	ws, err := h.workspaceRepo.FindByID(ctx, cmd.WorkspaceID)
	if err != nil {
		return err
	}

	// Only owner can delete workspace
	if ws.OwnerID != cmd.UserID {
		return workspace.ErrPermissionDenied
	}

	return h.workspaceRepo.Delete(ctx, cmd.WorkspaceID)
}
