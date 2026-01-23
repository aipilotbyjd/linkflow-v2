package websocket

import (
	"time"

	"github.com/google/uuid"
)

// Execution streaming events
const (
	EventExecutionProgress   = "execution.progress"
	EventExecutionNodeOutput = "execution.node_output"
	EventExecutionLog        = "execution.log"
	EventExecutionMetrics    = "execution.metrics"
	EventExecutionWaiting    = "execution.waiting"
	EventExecutionResumed    = "execution.resumed"
	EventApprovalRequired    = "execution.approval_required"
	EventApprovalReceived    = "execution.approval_received"
)

// ExecutionProgressData represents real-time execution progress
type ExecutionProgressData struct {
	ExecutionID     uuid.UUID `json:"execution_id"`
	WorkflowID      uuid.UUID `json:"workflow_id"`
	CurrentNodeID   string    `json:"current_node_id"`
	CurrentNodeName string    `json:"current_node_name"`
	CurrentNodeType string    `json:"current_node_type"`
	Progress        float64   `json:"progress"` // 0-100 percentage
	NodesCompleted  int       `json:"nodes_completed"`
	NodesTotal      int       `json:"nodes_total"`
	Status          string    `json:"status"`
	StartedAt       time.Time `json:"started_at"`
	ElapsedMs       int64     `json:"elapsed_ms"`
}

// NodeOutputData represents real-time node output for debugging
type NodeOutputData struct {
	ExecutionID uuid.UUID              `json:"execution_id"`
	NodeID      string                 `json:"node_id"`
	NodeName    string                 `json:"node_name"`
	NodeType    string                 `json:"node_type"`
	Input       map[string]interface{} `json:"input,omitempty"`
	Output      map[string]interface{} `json:"output,omitempty"`
	Error       *NodeErrorData         `json:"error,omitempty"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	DurationMs  int64                  `json:"duration_ms"`
	RetryCount  int                    `json:"retry_count,omitempty"`
}

// NodeErrorData represents error information for a node
type NodeErrorData struct {
	Message    string `json:"message"`
	Type       string `json:"type"`
	Code       string `json:"code,omitempty"`
	Retryable  bool   `json:"retryable"`
	StackTrace string `json:"stack_trace,omitempty"`
}

// ExecutionLogData represents execution log entry
type ExecutionLogData struct {
	ExecutionID uuid.UUID   `json:"execution_id"`
	NodeID      string      `json:"node_id,omitempty"`
	Level       string      `json:"level"` // debug, info, warn, error
	Message     string      `json:"message"`
	Data        interface{} `json:"data,omitempty"`
	Timestamp   time.Time   `json:"timestamp"`
}

// ExecutionMetricsData represents real-time execution metrics
type ExecutionMetricsData struct {
	ExecutionID   uuid.UUID `json:"execution_id"`
	MemoryUsageMB float64   `json:"memory_usage_mb"`
	CPUPercent    float64   `json:"cpu_percent"`
	DataProcessed int64     `json:"data_processed_bytes"`
	APICallsCount int       `json:"api_calls_count"`
	CreditsUsed   float64   `json:"credits_used"`
	EstimatedCost float64   `json:"estimated_cost_usd"`
}

// ApprovalRequiredData represents approval request notification
type ApprovalRequiredData struct {
	ExecutionID uuid.UUID   `json:"execution_id"`
	WorkflowID  uuid.UUID   `json:"workflow_id"`
	ApprovalID  string      `json:"approval_id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Approvers   []string    `json:"approvers"`
	ResumeURL   string      `json:"resume_url"`
	TimeoutAt   time.Time   `json:"timeout_at"`
	NodeID      string      `json:"node_id"`
	NodeName    string      `json:"node_name"`
	Data        interface{} `json:"data,omitempty"`
}

// ExecutionStreamService handles execution streaming
type ExecutionStreamService struct {
	hub *Hub
}

// NewExecutionStreamService creates a new execution stream service
func NewExecutionStreamService(hub *Hub) *ExecutionStreamService {
	return &ExecutionStreamService{hub: hub}
}

// SendProgress sends execution progress update
func (s *ExecutionStreamService) SendProgress(workspaceID uuid.UUID, data ExecutionProgressData) {
	s.hub.BroadcastToWorkspace(workspaceID, EventExecutionProgress, data)
}

// SendNodeOutput sends node output for real-time debugging
func (s *ExecutionStreamService) SendNodeOutput(workspaceID uuid.UUID, data NodeOutputData) {
	s.hub.BroadcastToWorkspace(workspaceID, EventExecutionNodeOutput, data)
}

// SendLog sends execution log entry
func (s *ExecutionStreamService) SendLog(workspaceID uuid.UUID, data ExecutionLogData) {
	s.hub.BroadcastToWorkspace(workspaceID, EventExecutionLog, data)
}

// SendMetrics sends execution metrics
func (s *ExecutionStreamService) SendMetrics(workspaceID uuid.UUID, data ExecutionMetricsData) {
	s.hub.BroadcastToWorkspace(workspaceID, EventExecutionMetrics, data)
}

// SendApprovalRequired notifies about approval requirement
func (s *ExecutionStreamService) SendApprovalRequired(workspaceID uuid.UUID, data ApprovalRequiredData) {
	s.hub.BroadcastToWorkspace(workspaceID, EventApprovalRequired, data)
}

// SendExecutionWaiting notifies that execution is waiting
func (s *ExecutionStreamService) SendExecutionWaiting(workspaceID uuid.UUID, executionID uuid.UUID, reason string, data interface{}) {
	s.hub.BroadcastToWorkspace(workspaceID, EventExecutionWaiting, map[string]interface{}{
		"execution_id": executionID,
		"reason":       reason,
		"data":         data,
		"timestamp":    time.Now(),
	})
}

// SendExecutionResumed notifies that execution was resumed
func (s *ExecutionStreamService) SendExecutionResumed(workspaceID uuid.UUID, executionID uuid.UUID, resumedBy string) {
	s.hub.BroadcastToWorkspace(workspaceID, EventExecutionResumed, map[string]interface{}{
		"execution_id": executionID,
		"resumed_by":   resumedBy,
		"timestamp":    time.Now(),
	})
}

// BroadcastToExecution sends event to all clients watching a specific execution
func (s *ExecutionStreamService) BroadcastToExecution(workspaceID, executionID uuid.UUID, event string, data interface{}) {
	s.hub.BroadcastToWorkspace(workspaceID, event, map[string]interface{}{
		"execution_id": executionID,
		"data":         data,
		"timestamp":    time.Now().UnixMilli(),
	})
}
