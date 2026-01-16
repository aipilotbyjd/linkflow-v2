package execution

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
)

type CancelExecutionCommand struct {
	ExecutionID uuid.UUID
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
	Reason      string
}

type CancelExecutionHandler struct {
	executionRepo execution.Repository
}

func NewCancelExecutionHandler(executionRepo execution.Repository) *CancelExecutionHandler {
	return &CancelExecutionHandler{executionRepo: executionRepo}
}

func (h *CancelExecutionHandler) Handle(ctx context.Context, cmd CancelExecutionCommand) (*execution.Execution, error) {
	exec, err := h.executionRepo.FindByID(ctx, cmd.ExecutionID)
	if err != nil {
		return nil, err
	}

	// Verify workspace
	if exec.WorkspaceID != cmd.WorkspaceID {
		return nil, execution.ErrExecutionNotFound
	}

	// Can only cancel running or queued executions
	if exec.Status != execution.StatusRunning && exec.Status != execution.StatusQueued {
		return nil, execution.ErrCannotCancel
	}

	// Cancel the execution
	exec.Cancel()
	if cmd.Reason != "" {
		exec.ErrorMessage = &cmd.Reason
	}

	if err := h.executionRepo.Update(ctx, exec); err != nil {
		return nil, err
	}

	return exec, nil
}
