package events

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Publisher struct {
	redis *redis.Client
}

func NewPublisher(redis *redis.Client) *Publisher {
	return &Publisher{redis: redis}
}

type EventType string

const (
	EventExecutionStarted   EventType = "execution.started"
	EventExecutionCompleted EventType = "execution.completed"
	EventExecutionFailed    EventType = "execution.failed"
	EventExecutionCancelled EventType = "execution.cancelled"
	EventNodeStarted        EventType = "node.started"
	EventNodeCompleted      EventType = "node.completed"
	EventNodeFailed         EventType = "node.failed"
	EventWorkflowActivated  EventType = "workflow.activated"
	EventWorkflowDeactivated EventType = "workflow.deactivated"
)

type Event struct {
	Type        EventType              `json:"type"`
	WorkspaceID uuid.UUID              `json:"workspace_id"`
	WorkflowID  uuid.UUID              `json:"workflow_id,omitempty"`
	ExecutionID uuid.UUID              `json:"execution_id,omitempty"`
	NodeID      string                 `json:"node_id,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

func (p *Publisher) Publish(ctx context.Context, event *Event) error {
	event.Timestamp = time.Now()

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	channel := "workspace:" + event.WorkspaceID.String()
	return p.redis.Publish(ctx, channel, data).Err()
}

func (p *Publisher) ExecutionStarted(ctx context.Context, workspaceID, workflowID, executionID uuid.UUID, triggerType string) error {
	return p.Publish(ctx, &Event{
		Type:        EventExecutionStarted,
		WorkspaceID: workspaceID,
		WorkflowID:  workflowID,
		ExecutionID: executionID,
		Data: map[string]interface{}{
			"trigger_type": triggerType,
			"status":       "running",
		},
	})
}

func (p *Publisher) ExecutionCompleted(ctx context.Context, workspaceID, workflowID, executionID uuid.UUID, durationMs int64, nodesCompleted int) error {
	return p.Publish(ctx, &Event{
		Type:        EventExecutionCompleted,
		WorkspaceID: workspaceID,
		WorkflowID:  workflowID,
		ExecutionID: executionID,
		Data: map[string]interface{}{
			"status":          "completed",
			"duration_ms":     durationMs,
			"nodes_completed": nodesCompleted,
		},
	})
}

func (p *Publisher) ExecutionFailed(ctx context.Context, workspaceID, workflowID, executionID uuid.UUID, errorMsg string, errorNodeID *string) error {
	data := map[string]interface{}{
		"status": "failed",
		"error":  errorMsg,
	}
	if errorNodeID != nil {
		data["error_node_id"] = *errorNodeID
	}

	return p.Publish(ctx, &Event{
		Type:        EventExecutionFailed,
		WorkspaceID: workspaceID,
		WorkflowID:  workflowID,
		ExecutionID: executionID,
		Data:        data,
	})
}

func (p *Publisher) NodeStarted(ctx context.Context, workspaceID, workflowID, executionID uuid.UUID, nodeID, nodeType, nodeName string) error {
	return p.Publish(ctx, &Event{
		Type:        EventNodeStarted,
		WorkspaceID: workspaceID,
		WorkflowID:  workflowID,
		ExecutionID: executionID,
		NodeID:      nodeID,
		Data: map[string]interface{}{
			"node_type": nodeType,
			"node_name": nodeName,
			"status":    "running",
		},
	})
}

func (p *Publisher) NodeCompleted(ctx context.Context, workspaceID, workflowID, executionID uuid.UUID, nodeID string, durationMs int, outputPreview interface{}) error {
	return p.Publish(ctx, &Event{
		Type:        EventNodeCompleted,
		WorkspaceID: workspaceID,
		WorkflowID:  workflowID,
		ExecutionID: executionID,
		NodeID:      nodeID,
		Data: map[string]interface{}{
			"status":         "completed",
			"duration_ms":    durationMs,
			"output_preview": outputPreview,
		},
	})
}

func (p *Publisher) NodeFailed(ctx context.Context, workspaceID, workflowID, executionID uuid.UUID, nodeID, errorMsg string) error {
	return p.Publish(ctx, &Event{
		Type:        EventNodeFailed,
		WorkspaceID: workspaceID,
		WorkflowID:  workflowID,
		ExecutionID: executionID,
		NodeID:      nodeID,
		Data: map[string]interface{}{
			"status": "failed",
			"error":  errorMsg,
		},
	})
}

func (p *Publisher) WorkflowActivated(ctx context.Context, workspaceID, workflowID uuid.UUID) error {
	return p.Publish(ctx, &Event{
		Type:        EventWorkflowActivated,
		WorkspaceID: workspaceID,
		WorkflowID:  workflowID,
		Data: map[string]interface{}{
			"status": "active",
		},
	})
}

func (p *Publisher) WorkflowDeactivated(ctx context.Context, workspaceID, workflowID uuid.UUID) error {
	return p.Publish(ctx, &Event{
		Type:        EventWorkflowDeactivated,
		WorkspaceID: workspaceID,
		WorkflowID:  workflowID,
		Data: map[string]interface{}{
			"status": "inactive",
		},
	})
}
