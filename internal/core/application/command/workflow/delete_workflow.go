package workflow

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
)

type DeleteWorkflowCommand struct {
	WorkflowID  uuid.UUID
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
}

type DeleteWorkflowHandler struct {
	workflowRepo workflow.Repository
}

func NewDeleteWorkflowHandler(workflowRepo workflow.Repository) *DeleteWorkflowHandler {
	return &DeleteWorkflowHandler{workflowRepo: workflowRepo}
}

func (h *DeleteWorkflowHandler) Handle(ctx context.Context, cmd DeleteWorkflowCommand) error {
	wf, err := h.workflowRepo.FindByID(ctx, cmd.WorkflowID)
	if err != nil {
		return err
	}

	// Verify workspace
	if wf.WorkspaceID != cmd.WorkspaceID {
		return workflow.ErrWorkflowNotFound
	}

	// Cannot delete active workflow
	if wf.Status == workflow.StatusActive {
		return workflow.ErrCannotDeleteActive
	}

	return h.workflowRepo.Delete(ctx, cmd.WorkflowID)
}
