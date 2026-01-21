package replay

import (
	"time"

	"github.com/google/uuid"
)

// ReplaySession represents a replay/time-travel debugging session
type ReplaySession struct {
	ID              uuid.UUID              `json:"id"`
	WorkspaceID     uuid.UUID              `json:"workspace_id"`
	OriginalExecID  uuid.UUID              `json:"original_execution_id"`
	WorkflowID      uuid.UUID              `json:"workflow_id"`
	UserID          uuid.UUID              `json:"user_id"`
	Status          ReplayStatus           `json:"status"`
	Mode            ReplayMode             `json:"mode"`
	Options         ReplayOptions          `json:"options"`
	CurrentNodeID   *string                `json:"current_node_id,omitempty"`
	Breakpoints     []string               `json:"breakpoints"`
	ModifiedInputs  map[string]interface{} `json:"modified_inputs,omitempty"`
	NewExecutionID  *uuid.UUID             `json:"new_execution_id,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	StartedAt       *time.Time             `json:"started_at,omitempty"`
	CompletedAt     *time.Time             `json:"completed_at,omitempty"`
}

// ReplayStatus represents the status of a replay session
type ReplayStatus string

const (
	ReplayStatusPending   ReplayStatus = "pending"
	ReplayStatusRunning   ReplayStatus = "running"
	ReplayStatusPaused    ReplayStatus = "paused"
	ReplayStatusCompleted ReplayStatus = "completed"
	ReplayStatusFailed    ReplayStatus = "failed"
	ReplayStatusCancelled ReplayStatus = "cancelled"
)

// ReplayMode represents the replay mode
type ReplayMode string

const (
	ReplayModeFullReplay    ReplayMode = "full"        // Replay entire execution
	ReplayModeFromNode      ReplayMode = "from_node"   // Start from specific node
	ReplayModeStepByStep    ReplayMode = "step"        // Execute one node at a time
	ReplayModeToBreakpoint  ReplayMode = "breakpoint"  // Run until breakpoint
)

// ReplayOptions configures replay behavior
type ReplayOptions struct {
	StartFromNodeID  *string                `json:"start_from_node_id,omitempty"`
	EndAtNodeID      *string                `json:"end_at_node_id,omitempty"`
	SkipNodes        []string               `json:"skip_nodes,omitempty"`
	ModifyInputs     map[string]interface{} `json:"modify_inputs,omitempty"`
	ModifyNodeParams map[string]map[string]interface{} `json:"modify_node_params,omitempty"`
	UseOriginalCreds bool                   `json:"use_original_creds"`
	CaptureSnapshots bool                   `json:"capture_snapshots"`
	MaxDuration      int                    `json:"max_duration_seconds,omitempty"`
}

// EventLog represents a captured event for replay
type EventLog struct {
	ID           uuid.UUID              `json:"id"`
	ExecutionID  uuid.UUID              `json:"execution_id"`
	WorkflowID   uuid.UUID              `json:"workflow_id"`
	WorkspaceID  uuid.UUID              `json:"workspace_id"`
	EventType    EventType              `json:"event_type"`
	NodeID       *string                `json:"node_id,omitempty"`
	NodeType     *string                `json:"node_type,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
	Data         map[string]interface{} `json:"data"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// EventType represents the type of event
type EventType string

const (
	EventTypeExecutionStarted   EventType = "execution.started"
	EventTypeExecutionCompleted EventType = "execution.completed"
	EventTypeExecutionFailed    EventType = "execution.failed"
	EventTypeNodeStarted        EventType = "node.started"
	EventTypeNodeCompleted      EventType = "node.completed"
	EventTypeNodeFailed         EventType = "node.failed"
	EventTypeNodeSkipped        EventType = "node.skipped"
	EventTypeDataTransfer       EventType = "data.transfer"
	EventTypeWebhookReceived    EventType = "webhook.received"
	EventTypeAPICall            EventType = "api.call"
	EventTypeAPIResponse        EventType = "api.response"
)

// NewReplaySession creates a new replay session
func NewReplaySession(workspaceID, originalExecID, workflowID, userID uuid.UUID, mode ReplayMode, opts ReplayOptions) *ReplaySession {
	return &ReplaySession{
		ID:             uuid.New(),
		WorkspaceID:    workspaceID,
		OriginalExecID: originalExecID,
		WorkflowID:     workflowID,
		UserID:         userID,
		Status:         ReplayStatusPending,
		Mode:           mode,
		Options:        opts,
		Breakpoints:    []string{},
		CreatedAt:      time.Now(),
	}
}

// Start starts the replay session
func (s *ReplaySession) Start() {
	s.Status = ReplayStatusRunning
	now := time.Now()
	s.StartedAt = &now
}

// Pause pauses the replay session
func (s *ReplaySession) Pause(nodeID string) {
	s.Status = ReplayStatusPaused
	s.CurrentNodeID = &nodeID
}

// Resume resumes the replay session
func (s *ReplaySession) Resume() {
	s.Status = ReplayStatusRunning
}

// Complete marks the session as completed
func (s *ReplaySession) Complete(newExecID uuid.UUID) {
	s.Status = ReplayStatusCompleted
	s.NewExecutionID = &newExecID
	now := time.Now()
	s.CompletedAt = &now
}

// Fail marks the session as failed
func (s *ReplaySession) Fail() {
	s.Status = ReplayStatusFailed
	now := time.Now()
	s.CompletedAt = &now
}

// AddBreakpoint adds a breakpoint at a node
func (s *ReplaySession) AddBreakpoint(nodeID string) {
	for _, bp := range s.Breakpoints {
		if bp == nodeID {
			return
		}
	}
	s.Breakpoints = append(s.Breakpoints, nodeID)
}

// RemoveBreakpoint removes a breakpoint
func (s *ReplaySession) RemoveBreakpoint(nodeID string) {
	for i, bp := range s.Breakpoints {
		if bp == nodeID {
			s.Breakpoints = append(s.Breakpoints[:i], s.Breakpoints[i+1:]...)
			return
		}
	}
}

// HasBreakpoint checks if a node has a breakpoint
func (s *ReplaySession) HasBreakpoint(nodeID string) bool {
	for _, bp := range s.Breakpoints {
		if bp == nodeID {
			return true
		}
	}
	return false
}

// ShouldSkipNode checks if a node should be skipped
func (s *ReplaySession) ShouldSkipNode(nodeID string) bool {
	for _, skip := range s.Options.SkipNodes {
		if skip == nodeID {
			return true
		}
	}
	return false
}
