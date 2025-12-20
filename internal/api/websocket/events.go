package websocket

import "time"

type EventType string

const (
	EventExecutionStarted   EventType = "execution.started"
	EventExecutionCompleted EventType = "execution.completed"
	EventExecutionFailed    EventType = "execution.failed"
	EventExecutionCancelled EventType = "execution.cancelled"
	EventExecutionProgress  EventType = "execution.progress"
	EventNodeStarted        EventType = "node.started"
	EventNodeCompleted      EventType = "node.completed"
	EventNodeFailed         EventType = "node.failed"
	EventWorkflowUpdated    EventType = "workflow.updated"
	EventWorkflowActivated  EventType = "workflow.activated"
	EventWorkflowDeactivated EventType = "workflow.deactivated"
)

type Event struct {
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

func NewEvent(eventType EventType, data map[string]interface{}) *Event {
	return &Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}
}

func ExecutionStartedEvent(executionID, workflowID string) *Event {
	return NewEvent(EventExecutionStarted, map[string]interface{}{
		"execution_id": executionID,
		"workflow_id":  workflowID,
	})
}

func ExecutionCompletedEvent(executionID, workflowID string, output interface{}) *Event {
	return NewEvent(EventExecutionCompleted, map[string]interface{}{
		"execution_id": executionID,
		"workflow_id":  workflowID,
		"output":       output,
	})
}

func ExecutionFailedEvent(executionID, workflowID, errorMessage string, errorNodeID *string) *Event {
	data := map[string]interface{}{
		"execution_id":  executionID,
		"workflow_id":   workflowID,
		"error_message": errorMessage,
	}
	if errorNodeID != nil {
		data["error_node_id"] = *errorNodeID
	}
	return NewEvent(EventExecutionFailed, data)
}

func ExecutionProgressEvent(executionID string, nodesCompleted, nodesTotal int) *Event {
	return NewEvent(EventExecutionProgress, map[string]interface{}{
		"execution_id":    executionID,
		"nodes_completed": nodesCompleted,
		"nodes_total":     nodesTotal,
	})
}

func NodeStartedEvent(executionID, nodeID, nodeType string) *Event {
	return NewEvent(EventNodeStarted, map[string]interface{}{
		"execution_id": executionID,
		"node_id":      nodeID,
		"node_type":    nodeType,
	})
}

func NodeCompletedEvent(executionID, nodeID string, output interface{}, durationMs int) *Event {
	return NewEvent(EventNodeCompleted, map[string]interface{}{
		"execution_id": executionID,
		"node_id":      nodeID,
		"output":       output,
		"duration_ms":  durationMs,
	})
}

func NodeFailedEvent(executionID, nodeID, errorMessage string) *Event {
	return NewEvent(EventNodeFailed, map[string]interface{}{
		"execution_id":  executionID,
		"node_id":       nodeID,
		"error_message": errorMessage,
	})
}
