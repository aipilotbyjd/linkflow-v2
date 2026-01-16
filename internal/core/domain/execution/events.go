package execution

import (
	"time"

	"github.com/google/uuid"
)

// Event types
const (
	EventExecutionStarted   = "execution.started"
	EventExecutionCompleted = "execution.completed"
	EventExecutionFailed    = "execution.failed"
	EventExecutionCancelled = "execution.cancelled"
	EventNodeStarted        = "execution.node_started"
	EventNodeCompleted      = "execution.node_completed"
	EventNodeFailed         = "execution.node_failed"
	EventNodeSkipped        = "execution.node_skipped"
)

// ExecutionStarted event
type ExecutionStarted struct {
	ExecutionID uuid.UUID `json:"execution_id"`
	WorkflowID  uuid.UUID `json:"workflow_id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	TriggerType string    `json:"trigger_type"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e ExecutionStarted) EventType() string    { return EventExecutionStarted }
func (e ExecutionStarted) AggregateID() string  { return e.ExecutionID.String() }
func (e ExecutionStarted) OccurredAt() time.Time { return e.Timestamp }

// ExecutionCompleted event
type ExecutionCompleted struct {
	ExecutionID uuid.UUID `json:"execution_id"`
	WorkflowID  uuid.UUID `json:"workflow_id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	DurationMs  int64     `json:"duration_ms"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e ExecutionCompleted) EventType() string    { return EventExecutionCompleted }
func (e ExecutionCompleted) AggregateID() string  { return e.ExecutionID.String() }
func (e ExecutionCompleted) OccurredAt() time.Time { return e.Timestamp }

// ExecutionFailed event
type ExecutionFailed struct {
	ExecutionID  uuid.UUID `json:"execution_id"`
	WorkflowID   uuid.UUID `json:"workflow_id"`
	WorkspaceID  uuid.UUID `json:"workspace_id"`
	ErrorMessage string    `json:"error_message"`
	FailedNodeID string    `json:"failed_node_id,omitempty"`
	DurationMs   int64     `json:"duration_ms"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e ExecutionFailed) EventType() string    { return EventExecutionFailed }
func (e ExecutionFailed) AggregateID() string  { return e.ExecutionID.String() }
func (e ExecutionFailed) OccurredAt() time.Time { return e.Timestamp }

// ExecutionCancelled event
type ExecutionCancelled struct {
	ExecutionID uuid.UUID `json:"execution_id"`
	WorkflowID  uuid.UUID `json:"workflow_id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	CancelledBy uuid.UUID `json:"cancelled_by,omitempty"`
	Reason      string    `json:"reason,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e ExecutionCancelled) EventType() string    { return EventExecutionCancelled }
func (e ExecutionCancelled) AggregateID() string  { return e.ExecutionID.String() }
func (e ExecutionCancelled) OccurredAt() time.Time { return e.Timestamp }

// NodeStarted event
type NodeStarted struct {
	ExecutionID uuid.UUID `json:"execution_id"`
	NodeID      string    `json:"node_id"`
	NodeType    string    `json:"node_type"`
	NodeName    string    `json:"node_name"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e NodeStarted) EventType() string    { return EventNodeStarted }
func (e NodeStarted) AggregateID() string  { return e.ExecutionID.String() }
func (e NodeStarted) OccurredAt() time.Time { return e.Timestamp }

// NodeCompleted event
type NodeCompleted struct {
	ExecutionID uuid.UUID `json:"execution_id"`
	NodeID      string    `json:"node_id"`
	NodeType    string    `json:"node_type"`
	DurationMs  int64     `json:"duration_ms"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e NodeCompleted) EventType() string    { return EventNodeCompleted }
func (e NodeCompleted) AggregateID() string  { return e.ExecutionID.String() }
func (e NodeCompleted) OccurredAt() time.Time { return e.Timestamp }

// NodeFailed event
type NodeFailed struct {
	ExecutionID  uuid.UUID `json:"execution_id"`
	NodeID       string    `json:"node_id"`
	NodeType     string    `json:"node_type"`
	ErrorMessage string    `json:"error_message"`
	RetryCount   int       `json:"retry_count"`
	DurationMs   int64     `json:"duration_ms"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e NodeFailed) EventType() string    { return EventNodeFailed }
func (e NodeFailed) AggregateID() string  { return e.ExecutionID.String() }
func (e NodeFailed) OccurredAt() time.Time { return e.Timestamp }

// NodeSkipped event
type NodeSkipped struct {
	ExecutionID uuid.UUID `json:"execution_id"`
	NodeID      string    `json:"node_id"`
	NodeType    string    `json:"node_type"`
	Reason      string    `json:"reason"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e NodeSkipped) EventType() string    { return EventNodeSkipped }
func (e NodeSkipped) AggregateID() string  { return e.ExecutionID.String() }
func (e NodeSkipped) OccurredAt() time.Time { return e.Timestamp }
