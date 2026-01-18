package asynq

import (
	"context"

	"github.com/hibiken/asynq"
)

// TaskQueueAdapter adapts the Asynq client to the TaskQueue interface
type TaskQueueAdapter struct {
	client *Client
}

// NewTaskQueueAdapter creates a new task queue adapter
func NewTaskQueueAdapter(client *Client) *TaskQueueAdapter {
	return &TaskQueueAdapter{client: client}
}

// EnqueueWorkflowExecution enqueues a workflow execution task
func (a *TaskQueueAdapter) EnqueueWorkflowExecution(ctx context.Context, payload interface{}) error {
	p, ok := payload.(map[string]interface{})
	if !ok {
		return nil
	}

	execPayload := ExecuteWorkflowPayload{
		ExecutionID: getString(p, "execution_id"),
		WorkflowID:  getString(p, "workflow_id"),
		WorkspaceID: getString(p, "workspace_id"),
		TriggerType: getString(p, "trigger_type"),
	}

	if inputData, ok := p["input_data"].(map[string]interface{}); ok {
		execPayload.InputData = inputData
	}

	_, err := a.client.EnqueueWorkflowExecution(ctx, execPayload, asynq.MaxRetry(3))
	return err
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
