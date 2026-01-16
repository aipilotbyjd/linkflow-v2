package execution

import (
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// NodeExecution represents a single node execution
type NodeExecution struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ExecutionID  uuid.UUID  `gorm:"type:uuid;index;not null" json:"execution_id"`
	NodeID       string     `gorm:"size:100;not null" json:"node_id"`
	NodeType     string     `gorm:"size:50;not null" json:"node_type"`
	NodeName     *string    `gorm:"size:255" json:"node_name,omitempty"`
	Status       NodeStatus `gorm:"size:20;not null;default:pending;index" json:"status"`
	InputData    types.JSON `gorm:"type:jsonb" json:"input_data,omitempty"`
	OutputData   types.JSON `gorm:"type:jsonb" json:"output_data,omitempty"`
	ErrorMessage *string    `gorm:"type:text" json:"error_message,omitempty"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	DurationMs   *int       `json:"duration_ms,omitempty"`
	RetryCount   int        `gorm:"default:0" json:"retry_count"`
	CreatedAt    time.Time  `json:"created_at"`

	Execution Execution `gorm:"foreignKey:ExecutionID" json:"-"`
}

func (NodeExecution) TableName() string {
	return "node_executions"
}

// NewNodeExecution creates a new node execution
func NewNodeExecution(executionID uuid.UUID, nodeID, nodeType string, nodeName *string) *NodeExecution {
	return &NodeExecution{
		ID:          uuid.New(),
		ExecutionID: executionID,
		NodeID:      nodeID,
		NodeType:    nodeType,
		NodeName:    nodeName,
		Status:      NodeStatusPending,
		CreatedAt:   time.Now(),
	}
}

// Start marks node as started
func (n *NodeExecution) Start() {
	n.Status = NodeStatusRunning
	now := time.Now()
	n.StartedAt = &now
}

// Complete marks node as completed
func (n *NodeExecution) Complete(outputData types.JSON) {
	n.Status = NodeStatusCompleted
	now := time.Now()
	n.CompletedAt = &now
	n.OutputData = outputData
	if n.StartedAt != nil {
		durationMs := int(now.Sub(*n.StartedAt).Milliseconds())
		n.DurationMs = &durationMs
	}
}

// Fail marks node as failed
func (n *NodeExecution) Fail(errorMsg string) {
	n.Status = NodeStatusFailed
	now := time.Now()
	n.CompletedAt = &now
	n.ErrorMessage = &errorMsg
	if n.StartedAt != nil {
		durationMs := int(now.Sub(*n.StartedAt).Milliseconds())
		n.DurationMs = &durationMs
	}
}

// Skip marks node as skipped
func (n *NodeExecution) Skip() {
	n.Status = NodeStatusSkipped
	now := time.Now()
	n.CompletedAt = &now
}

// SetInputData sets the input data
func (n *NodeExecution) SetInputData(data types.JSON) {
	n.InputData = data
}

// IncrementRetry increments retry count
func (n *NodeExecution) IncrementRetry() {
	n.RetryCount++
}

// Duration returns node execution duration
func (n *NodeExecution) Duration() time.Duration {
	if n.StartedAt == nil {
		return 0
	}
	endTime := time.Now()
	if n.CompletedAt != nil {
		endTime = *n.CompletedAt
	}
	return endTime.Sub(*n.StartedAt)
}

// Log represents an execution log entry
type Log struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ExecutionID uuid.UUID  `gorm:"type:uuid;index;not null" json:"execution_id"`
	NodeID      *string    `gorm:"size:100" json:"node_id,omitempty"`
	Level       LogLevel   `gorm:"size:10;not null" json:"level"`
	Message     string     `gorm:"type:text;not null" json:"message"`
	Data        types.JSON `gorm:"type:jsonb" json:"data,omitempty"`
	Timestamp   time.Time  `gorm:"not null;index" json:"timestamp"`

	Execution Execution `gorm:"foreignKey:ExecutionID" json:"-"`
}

func (Log) TableName() string {
	return "execution_logs"
}

// NewLog creates a new log entry
func NewLog(executionID uuid.UUID, level LogLevel, message string) *Log {
	return &Log{
		ID:          uuid.New(),
		ExecutionID: executionID,
		Level:       level,
		Message:     message,
		Timestamp:   time.Now(),
	}
}

// ForNode associates the log with a specific node
func (l *Log) ForNode(nodeID string) *Log {
	l.NodeID = &nodeID
	return l
}

// WithData adds data to the log
func (l *Log) WithData(data types.JSON) *Log {
	l.Data = data
	return l
}
