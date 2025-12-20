package events

import (
	"time"

	"github.com/google/uuid"
)

// LogLevel represents log severity
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       LogLevel               `json:"level"`
	Message     string                 `json:"message"`
	ExecutionID uuid.UUID              `json:"execution_id"`
	WorkflowID  uuid.UUID              `json:"workflow_id"`
	NodeID      string                 `json:"node_id,omitempty"`
	NodeType    string                 `json:"node_type,omitempty"`
	Duration    int64                  `json:"duration_ms,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ProgressEvent represents an execution progress update
type ProgressEvent struct {
	ExecutionID    uuid.UUID `json:"execution_id"`
	WorkflowID     uuid.UUID `json:"workflow_id"`
	TotalNodes     int       `json:"total_nodes"`
	CompletedNodes int       `json:"completed_nodes"`
	Percentage     int       `json:"percentage"`
	CurrentNode    string    `json:"current_node,omitempty"`
	Status         string    `json:"status"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// NodeStatusEvent represents a node status change
type NodeStatusEvent struct {
	ExecutionID uuid.UUID              `json:"execution_id"`
	WorkflowID  uuid.UUID              `json:"workflow_id"`
	NodeID      string                 `json:"node_id"`
	NodeType    string                 `json:"node_type"`
	NodeName    string                 `json:"node_name"`
	Status      string                 `json:"status"`
	Duration    int64                  `json:"duration_ms,omitempty"`
	Output      map[string]interface{} `json:"output,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// WorkflowStatusEvent represents a workflow status change
type WorkflowStatusEvent struct {
	ExecutionID uuid.UUID              `json:"execution_id"`
	WorkflowID  uuid.UUID              `json:"workflow_id"`
	WorkspaceID uuid.UUID              `json:"workspace_id"`
	Status      string                 `json:"status"`
	TriggerType string                 `json:"trigger_type,omitempty"`
	Duration    int64                  `json:"duration_ms,omitempty"`
	NodesCount  int                    `json:"nodes_count,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// MetricEvent represents a metric update
type MetricEvent struct {
	Name        string                 `json:"name"`
	Value       float64                `json:"value"`
	Unit        string                 `json:"unit,omitempty"`
	Tags        map[string]string      `json:"tags,omitempty"`
	WorkspaceID uuid.UUID              `json:"workspace_id,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// AlertEvent represents an alert notification
type AlertEvent struct {
	ExecutionID uuid.UUID              `json:"execution_id,omitempty"`
	WorkflowID  uuid.UUID              `json:"workflow_id,omitempty"`
	WorkspaceID uuid.UUID              `json:"workspace_id"`
	Level       string                 `json:"level"` // info, warning, error, critical
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	NodeID      string                 `json:"node_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// EventSubscription represents an event subscription
type EventSubscription struct {
	ID          string    `json:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	EventTypes  []string  `json:"event_types"`
	Channel     string    `json:"channel"`
	CreatedAt   time.Time `json:"created_at"`
}

// EventFilter for filtering events
type EventFilter struct {
	WorkspaceID  *uuid.UUID
	WorkflowID   *uuid.UUID
	ExecutionID  *uuid.UUID
	EventTypes   []EventType
	NodeID       string
	MinTimestamp *time.Time
	MaxTimestamp *time.Time
}

// Matches checks if an event matches the filter
func (f *EventFilter) Matches(event *Event) bool {
	if f.WorkspaceID != nil && event.WorkspaceID != *f.WorkspaceID {
		return false
	}
	if f.WorkflowID != nil && event.WorkflowID != *f.WorkflowID {
		return false
	}
	if f.ExecutionID != nil && event.ExecutionID != *f.ExecutionID {
		return false
	}
	if len(f.EventTypes) > 0 {
		found := false
		for _, t := range f.EventTypes {
			if t == event.Type {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if f.NodeID != "" && event.NodeID != f.NodeID {
		return false
	}
	if f.MinTimestamp != nil && event.Timestamp.Before(*f.MinTimestamp) {
		return false
	}
	if f.MaxTimestamp != nil && event.Timestamp.After(*f.MaxTimestamp) {
		return false
	}
	return true
}

// EventBatch represents a batch of events
type EventBatch struct {
	Events    []*Event `json:"events"`
	BatchID   string   `json:"batch_id"`
	Timestamp time.Time `json:"timestamp"`
}

// NewEventBatch creates a new event batch
func NewEventBatch(events []*Event) *EventBatch {
	return &EventBatch{
		Events:    events,
		BatchID:   uuid.New().String(),
		Timestamp: time.Now(),
	}
}
