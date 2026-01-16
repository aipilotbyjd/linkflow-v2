package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

// WorkflowExecutionPayload contains data for workflow execution task
type WorkflowExecutionPayload struct {
	ExecutionID string                 `json:"execution_id"`
	WorkflowID  string                 `json:"workflow_id"`
	WorkspaceID string                 `json:"workspace_id"`
	TriggerType string                 `json:"trigger_type"`
	TriggerData map[string]interface{} `json:"trigger_data,omitempty"`
	InputData   map[string]interface{} `json:"input_data,omitempty"`
}

// NewWorkflowExecutionTask creates a new workflow execution task
func NewWorkflowExecutionTask(payload WorkflowExecutionPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	return asynq.NewTask(
		TypeWorkflowExecution,
		data,
		asynq.MaxRetry(3),
		asynq.Timeout(30*time.Minute),
		asynq.Queue(QueueDefault),
	), nil
}

// WorkflowExecutionHandler handles workflow execution tasks
type WorkflowExecutionHandler struct {
	executor WorkflowExecutor
}

// WorkflowExecutor interface for executing workflows
type WorkflowExecutor interface {
	Execute(ctx context.Context, executionID, workflowID, workspaceID string, triggerType string, triggerData, inputData map[string]interface{}) error
}

// NewWorkflowExecutionHandler creates a new handler
func NewWorkflowExecutionHandler(executor WorkflowExecutor) *WorkflowExecutionHandler {
	return &WorkflowExecutionHandler{executor: executor}
}

// ProcessTask processes a workflow execution task
func (h *WorkflowExecutionHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	var payload WorkflowExecutionPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return h.executor.Execute(
		ctx,
		payload.ExecutionID,
		payload.WorkflowID,
		payload.WorkspaceID,
		payload.TriggerType,
		payload.TriggerData,
		payload.InputData,
	)
}
