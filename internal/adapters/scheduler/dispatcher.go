package scheduler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/linkflow-ai/linkflow/internal/core/domain/schedule"
)

const (
	TaskTypeExecuteWorkflow = "workflow:execute"
)

type ExecuteWorkflowPayload struct {
	WorkflowID  uuid.UUID              `json:"workflow_id"`
	WorkspaceID uuid.UUID              `json:"workspace_id"`
	ScheduleID  uuid.UUID              `json:"schedule_id"`
	TriggerType string                 `json:"trigger_type"`
	InputData   map[string]interface{} `json:"input_data,omitempty"`
}

type Dispatcher struct {
	client       *asynq.Client
	scheduleRepo schedule.Repository
}

func NewDispatcher(client *asynq.Client, scheduleRepo schedule.Repository) *Dispatcher {
	return &Dispatcher{
		client:       client,
		scheduleRepo: scheduleRepo,
	}
}

// Dispatch enqueues a workflow execution task for the given schedule
func (d *Dispatcher) Dispatch(ctx context.Context, sched *schedule.Schedule) error {
	payload := ExecuteWorkflowPayload{
		WorkflowID:  sched.WorkflowID,
		WorkspaceID: sched.WorkspaceID,
		ScheduleID:  sched.ID,
		TriggerType: "schedule",
		InputData:   sched.InputData,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TaskTypeExecuteWorkflow, data)

	_, err = d.client.Enqueue(task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	// Update schedule's last run time and calculate next run
	sched.MarkExecuted()
	if err := d.scheduleRepo.Update(ctx, sched); err != nil {
		// Log error but don't fail - task was already enqueued
	}

	return nil
}

// DispatchImmediate enqueues a workflow execution task immediately
func (d *Dispatcher) DispatchImmediate(ctx context.Context, workflowID, workspaceID uuid.UUID, inputData map[string]interface{}) error {
	payload := ExecuteWorkflowPayload{
		WorkflowID:  workflowID,
		WorkspaceID: workspaceID,
		TriggerType: "manual",
		InputData:   inputData,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TaskTypeExecuteWorkflow, data)

	_, err = d.client.Enqueue(task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	return nil
}
