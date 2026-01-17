package execution

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
	"github.com/linkflow-ai/linkflow/internal/shared/events"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// StartExecutionCommand represents the command to start a workflow execution
type StartExecutionCommand struct {
	WorkflowID  uuid.UUID
	WorkspaceID uuid.UUID
	TriggeredBy *uuid.UUID
	TriggerType string
	TriggerData types.JSON
	InputData   types.JSON
	Priority    int
}

// StartExecutionHandler handles starting workflow executions
type StartExecutionHandler struct {
	workflowRepo  workflow.Repository
	executionRepo execution.Repository
	eventBus      events.Bus
}

// NewStartExecutionHandler creates a new handler
func NewStartExecutionHandler(
	workflowRepo workflow.Repository,
	executionRepo execution.Repository,
	eventBus events.Bus,
) *StartExecutionHandler {
	return &StartExecutionHandler{
		workflowRepo:  workflowRepo,
		executionRepo: executionRepo,
		eventBus:      eventBus,
	}
}

// Handle executes the start execution command
func (h *StartExecutionHandler) Handle(ctx context.Context, cmd StartExecutionCommand) (*execution.Execution, error) {
	// Get workflow
	wf, err := h.workflowRepo.FindByID(ctx, cmd.WorkflowID)
	if err != nil {
		return nil, workflow.ErrWorkflowNotFound
	}

	// Validate workflow is active
	if !wf.IsActive() {
		return nil, execution.ErrWorkflowNotActive
	}

	// Create execution
	exec := execution.NewExecution(cmd.WorkflowID, cmd.WorkspaceID, wf.Version, cmd.TriggerType)

	if cmd.TriggeredBy != nil {
		exec.WithTriggeredBy(*cmd.TriggeredBy)
	}
	if cmd.TriggerData != nil {
		exec.WithTriggerData(cmd.TriggerData)
	}
	if cmd.InputData != nil {
		exec.WithInputData(cmd.InputData)
	}
	if cmd.Priority > 0 {
		exec.WithPriority(cmd.Priority)
	}

	// Set timeout from workflow settings
	exec.WithTimeout(wf.GetTimeoutSeconds())
	exec.WithMaxRetries(wf.GetMaxRetries())

	// Set total nodes count
	exec.SetNodesTotal(len(wf.Nodes))

	// Save
	if err := h.executionRepo.Create(ctx, exec); err != nil {
		return nil, fmt.Errorf("failed to create execution: %w", err)
	}

	// Increment workflow execution count
	_ = h.workflowRepo.IncrementExecutionCount(ctx, cmd.WorkflowID)

	// Publish event
	if h.eventBus != nil {
		event := events.ExecutionStarted{
			BaseEvent:   events.NewBaseEvent("execution.started", exec.ID, "execution"),
			ExecutionID: exec.ID,
			WorkflowID:  exec.WorkflowID,
			WorkspaceID: exec.WorkspaceID,
			TriggerType: exec.TriggerType,
		}
		_ = h.eventBus.Publish(ctx, event)
	}

	return exec, nil
}
