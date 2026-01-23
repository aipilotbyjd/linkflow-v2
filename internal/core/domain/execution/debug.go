package execution

import (
	"time"

	"github.com/google/uuid"
)

// NodeSnapshot represents the complete state of a node at execution time
type NodeSnapshot struct {
	ID             uuid.UUID              `json:"id"`
	ExecutionID    uuid.UUID              `json:"execution_id"`
	NodeID         string                 `json:"node_id"`
	NodeType       string                 `json:"node_type"`
	NodeName       string                 `json:"node_name"`
	SequenceNumber int                    `json:"sequence_number"`
	Input          map[string]interface{} `json:"input"`
	Output         map[string]interface{} `json:"output,omitempty"`
	Parameters     map[string]interface{} `json:"parameters"`
	CredentialID   *uuid.UUID             `json:"credential_id,omitempty"`
	Status         string                 `json:"status"`
	Error          *NodeError             `json:"error,omitempty"`
	StartedAt      time.Time              `json:"started_at"`
	CompletedAt    *time.Time             `json:"completed_at,omitempty"`
	DurationMs     int64                  `json:"duration_ms"`
	RetryCount     int                    `json:"retry_count"`
	MemoryUsageKB  int64                  `json:"memory_usage_kb,omitempty"`
}

// NodeError represents detailed error information
type NodeError struct {
	Message    string `json:"message"`
	Type       string `json:"type"`
	Code       string `json:"code,omitempty"`
	StackTrace string `json:"stack_trace,omitempty"`
	Retryable  bool   `json:"retryable"`
}

// ExecutionTimeline represents the complete timeline of an execution
type ExecutionTimeline struct {
	ExecutionID   uuid.UUID        `json:"execution_id"`
	WorkflowID    uuid.UUID        `json:"workflow_id"`
	WorkflowName  string           `json:"workflow_name"`
	TriggerType   string           `json:"trigger_type"`
	TriggerData   interface{}      `json:"trigger_data,omitempty"`
	Status        string           `json:"status"`
	StartedAt     time.Time        `json:"started_at"`
	CompletedAt   *time.Time       `json:"completed_at,omitempty"`
	TotalDuration int64            `json:"total_duration_ms"`
	Nodes         []NodeSnapshot   `json:"nodes"`
	DataFlow      []DataFlowEntry  `json:"data_flow"`
	Metrics       ExecutionMetrics `json:"metrics"`
}

// DataFlowEntry tracks data movement between nodes
type DataFlowEntry struct {
	FromNodeID string                 `json:"from_node_id"`
	ToNodeID   string                 `json:"to_node_id"`
	OutputPort string                 `json:"output_port"`
	InputPort  string                 `json:"input_port"`
	Data       map[string]interface{} `json:"data"`
	Timestamp  time.Time              `json:"timestamp"`
}

// ExecutionMetrics holds execution-level metrics
type ExecutionMetrics struct {
	TotalNodes        int     `json:"total_nodes"`
	CompletedNodes    int     `json:"completed_nodes"`
	FailedNodes       int     `json:"failed_nodes"`
	SkippedNodes      int     `json:"skipped_nodes"`
	TotalRetries      int     `json:"total_retries"`
	TotalAPICallsTime int64   `json:"total_api_calls_ms"`
	TotalDataBytes    int64   `json:"total_data_bytes"`
	CreditsUsed       float64 `json:"credits_used"`
}

// ReplayOptions configures how to replay an execution
type ReplayOptions struct {
	FromNodeID      string                 `json:"from_node_id,omitempty"`
	ToNodeID        string                 `json:"to_node_id,omitempty"`
	ModifyInput     map[string]interface{} `json:"modify_input,omitempty"`
	ModifyParams    map[string]interface{} `json:"modify_params,omitempty"`
	SkipNodes       []string               `json:"skip_nodes,omitempty"`
	BreakpointNodes []string               `json:"breakpoint_nodes,omitempty"`
	StepMode        bool                   `json:"step_mode"`
}

// DebugSession represents an active debugging session
type DebugSession struct {
	ID             uuid.UUID `json:"id"`
	ExecutionID    uuid.UUID `json:"execution_id"`
	UserID         uuid.UUID `json:"user_id"`
	Status         string    `json:"status"` // running, paused, stepping, completed
	CurrentNodeID  string    `json:"current_node_id,omitempty"`
	Breakpoints    []string  `json:"breakpoints"`
	WatchVariables []string  `json:"watch_variables"`
	StepMode       bool      `json:"step_mode"`
	CreatedAt      time.Time `json:"created_at"`
	LastActivity   time.Time `json:"last_activity"`
}

// NewNodeSnapshot creates a new node snapshot
func NewNodeSnapshot(executionID uuid.UUID, nodeID, nodeType, nodeName string, seqNum int) *NodeSnapshot {
	return &NodeSnapshot{
		ID:             uuid.New(),
		ExecutionID:    executionID,
		NodeID:         nodeID,
		NodeType:       nodeType,
		NodeName:       nodeName,
		SequenceNumber: seqNum,
		Status:         "pending",
		StartedAt:      time.Now(),
	}
}

// MarkStarted marks the node as started
func (s *NodeSnapshot) MarkStarted(input, params map[string]interface{}) {
	s.Status = "running"
	s.Input = input
	s.Parameters = params
	s.StartedAt = time.Now()
}

// MarkCompleted marks the node as completed
func (s *NodeSnapshot) MarkCompleted(output map[string]interface{}) {
	s.Status = "completed"
	s.Output = output
	now := time.Now()
	s.CompletedAt = &now
	s.DurationMs = now.Sub(s.StartedAt).Milliseconds()
}

// MarkFailed marks the node as failed
func (s *NodeSnapshot) MarkFailed(err NodeError) {
	s.Status = "failed"
	s.Error = &err
	now := time.Now()
	s.CompletedAt = &now
	s.DurationMs = now.Sub(s.StartedAt).Milliseconds()
}
