package execution

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
)

type RetryExecutionCommand struct {
	ExecutionID uuid.UUID
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
}

type RetryExecutionHandler struct {
	executionRepo execution.Repository
}

func NewRetryExecutionHandler(executionRepo execution.Repository) *RetryExecutionHandler {
	return &RetryExecutionHandler{executionRepo: executionRepo}
}

func (h *RetryExecutionHandler) Handle(ctx context.Context, cmd RetryExecutionCommand) (*execution.Execution, error) {
	original, err := h.executionRepo.FindByID(ctx, cmd.ExecutionID)
	if err != nil {
		return nil, err
	}

	// Verify workspace
	if original.WorkspaceID != cmd.WorkspaceID {
		return nil, execution.ErrExecutionNotFound
	}

	// Can only retry failed or cancelled executions
	if original.Status != execution.StatusFailed && original.Status != execution.StatusCancelled {
		return nil, execution.ErrCannotRetry
	}

	// Create new execution based on original
	newExec := execution.NewExecution(
		original.WorkflowID,
		original.WorkspaceID,
		original.WorkflowVersion,
		"retry",
	)
	newExec.TriggerData = original.TriggerData
	newExec.InputData = original.InputData
	newExec.ParentExecutionID = &original.ID

	if err := h.executionRepo.Create(ctx, newExec); err != nil {
		return nil, err
	}

	return newExec, nil
}
