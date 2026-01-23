package execution

import (
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// Execution entity (aggregate root)
type Execution struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkflowID        uuid.UUID  `gorm:"type:uuid;index;not null" json:"workflow_id"`
	WorkspaceID       uuid.UUID  `gorm:"type:uuid;index;not null" json:"workspace_id"`
	TriggeredBy       *uuid.UUID `gorm:"type:uuid" json:"triggered_by,omitempty"`
	WorkflowVersion   int        `gorm:"not null" json:"workflow_version"`
	Status            Status     `gorm:"size:20;not null;default:queued;index" json:"status"`
	TriggerType       string     `gorm:"size:20;not null" json:"trigger_type"`
	TriggerData       types.JSON `gorm:"type:jsonb" json:"trigger_data,omitempty"`
	InputData         types.JSON `gorm:"type:jsonb" json:"input_data,omitempty"`
	OutputData        types.JSON `gorm:"type:jsonb" json:"output_data,omitempty"`
	ErrorMessage      *string    `gorm:"type:text" json:"error_message,omitempty"`
	ErrorNodeID       *string    `gorm:"size:100" json:"error_node_id,omitempty"`
	QueuedAt          time.Time  `gorm:"default:now()" json:"queued_at"`
	StartedAt         *time.Time `json:"started_at,omitempty"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
	PausedAt          *time.Time `json:"paused_at,omitempty"`
	ResumedAt         *time.Time `json:"resumed_at,omitempty"`
	NodesTotal        int        `gorm:"default:0" json:"nodes_total"`
	NodesCompleted    int        `gorm:"default:0" json:"nodes_completed"`
	RetryCount        int        `gorm:"default:0" json:"retry_count"`
	MaxRetries        int        `gorm:"default:3" json:"max_retries"`
	Priority          int        `gorm:"default:5;index" json:"priority"`
	TimeoutSeconds    int        `gorm:"default:3600" json:"timeout_seconds"`
	ParentExecutionID *uuid.UUID `gorm:"type:uuid" json:"parent_execution_id,omitempty"`
	BatchID           *uuid.UUID `gorm:"type:uuid;index" json:"batch_id,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`

	NodeExecutions []NodeExecution `gorm:"foreignKey:ExecutionID" json:"-"`
}

func (Execution) TableName() string {
	return "executions"
}

// NewExecution creates a new execution
func NewExecution(workflowID, workspaceID uuid.UUID, version int, triggerType string) *Execution {
	now := time.Now()
	return &Execution{
		ID:              uuid.New(),
		WorkflowID:      workflowID,
		WorkspaceID:     workspaceID,
		WorkflowVersion: version,
		Status:          StatusQueued,
		TriggerType:     triggerType,
		QueuedAt:        now,
		Priority:        5,
		MaxRetries:      3,
		TimeoutSeconds:  3600,
		CreatedAt:       now,
	}
}

// GetWorkspaceID implements the WorkspaceOwned interface
func (e *Execution) GetWorkspaceID() uuid.UUID {
	return e.WorkspaceID
}

// WithTriggeredBy sets who triggered the execution
func (e *Execution) WithTriggeredBy(userID uuid.UUID) *Execution {
	e.TriggeredBy = &userID
	return e
}

// WithInputData sets the input data
func (e *Execution) WithInputData(data types.JSON) *Execution {
	e.InputData = data
	return e
}

// WithTriggerData sets the trigger data
func (e *Execution) WithTriggerData(data types.JSON) *Execution {
	e.TriggerData = data
	return e
}

// WithPriority sets the priority
func (e *Execution) WithPriority(priority int) *Execution {
	if priority < 1 {
		priority = 1
	}
	if priority > 10 {
		priority = 10
	}
	e.Priority = priority
	return e
}

// WithTimeout sets the timeout
func (e *Execution) WithTimeout(seconds int) *Execution {
	if seconds > 0 {
		e.TimeoutSeconds = seconds
	}
	return e
}

// WithMaxRetries sets the max retries
func (e *Execution) WithMaxRetries(retries int) *Execution {
	e.MaxRetries = retries
	return e
}

// WithParentExecution sets the parent execution for sub-workflows
func (e *Execution) WithParentExecution(parentID uuid.UUID) *Execution {
	e.ParentExecutionID = &parentID
	return e
}

// WithBatchID sets the batch ID for bulk executions
func (e *Execution) WithBatchID(batchID uuid.UUID) *Execution {
	e.BatchID = &batchID
	return e
}

// CanTransitionTo checks if transition to target status is valid
func (e *Execution) CanTransitionTo(target Status) bool {
	// Terminal states cannot transition to anything
	if e.Status.IsTerminal() {
		return false
	}

	switch target {
	case StatusRunning:
		return e.Status == StatusQueued || e.Status == StatusWaiting
	case StatusCompleted, StatusFailed, StatusTimeout:
		return e.Status == StatusRunning
	case StatusCancelled:
		return e.Status == StatusQueued || e.Status == StatusRunning || e.Status == StatusWaiting
	case StatusWaiting:
		return e.Status == StatusRunning
	default:
		return false
	}
}

// Start marks execution as started
func (e *Execution) Start() error {
	if !e.CanTransitionTo(StatusRunning) {
		return ErrInvalidStateTransition
	}
	e.Status = StatusRunning
	now := time.Now()
	// Only set StartedAt if it's the first start (not resume)
	if e.StartedAt == nil {
		e.StartedAt = &now
	}
	return nil
}

// Complete marks execution as completed
func (e *Execution) Complete(outputData types.JSON) error {
	if !e.CanTransitionTo(StatusCompleted) {
		return ErrInvalidStateTransition
	}
	e.Status = StatusCompleted
	now := time.Now()
	e.CompletedAt = &now
	e.OutputData = outputData
	return nil
}

// Fail marks execution as failed
func (e *Execution) Fail(errorMsg string, errorNodeID *string) error {
	if !e.CanTransitionTo(StatusFailed) {
		return ErrInvalidStateTransition
	}
	e.Status = StatusFailed
	now := time.Now()
	e.CompletedAt = &now
	e.ErrorMessage = &errorMsg
	e.ErrorNodeID = errorNodeID
	return nil
}

// Cancel marks execution as canceled
func (e *Execution) Cancel() error {
	if !e.CanTransitionTo(StatusCancelled) {
		return ErrInvalidStateTransition
	}
	e.Status = StatusCancelled
	now := time.Now()
	e.CompletedAt = &now
	return nil
}

// Timeout marks execution as timed out
func (e *Execution) Timeout() error {
	if !e.CanTransitionTo(StatusTimeout) {
		return ErrInvalidStateTransition
	}
	e.Status = StatusTimeout
	now := time.Now()
	e.CompletedAt = &now
	msg := "execution timed out"
	e.ErrorMessage = &msg
	return nil
}

// Pause marks execution as waiting (paused)
func (e *Execution) Pause() error {
	if !e.CanTransitionTo(StatusWaiting) {
		return ErrInvalidStateTransition
	}
	e.Status = StatusWaiting
	now := time.Now()
	e.PausedAt = &now
	return nil
}

// Resume resumes a paused execution
func (e *Execution) Resume() error {
	if !e.CanTransitionTo(StatusRunning) {
		return ErrInvalidStateTransition
	}
	e.Status = StatusRunning
	now := time.Now()
	e.ResumedAt = &now
	return nil
}

// IncrementRetry increments the retry count
func (e *Execution) IncrementRetry() {
	e.RetryCount++
}

// CanRetry checks if execution can be retried
func (e *Execution) CanRetry() bool {
	return e.RetryCount < e.MaxRetries
}

// SetNodesTotal sets the total number of nodes
func (e *Execution) SetNodesTotal(total int) {
	e.NodesTotal = total
}

// IncrementNodesCompleted increments completed nodes count
func (e *Execution) IncrementNodesCompleted() {
	e.NodesCompleted++
}

// IsFinished checks if execution is in a terminal state
func (e *Execution) IsFinished() bool {
	return e.Status == StatusCompleted ||
		e.Status == StatusFailed ||
		e.Status == StatusCancelled ||
		e.Status == StatusTimeout
}

// IsRunning checks if execution is running
func (e *Execution) IsRunning() bool {
	return e.Status == StatusRunning
}

// IsPaused checks if execution is paused
func (e *Execution) IsPaused() bool {
	return e.Status == StatusWaiting
}

// Duration returns execution duration
func (e *Execution) Duration() time.Duration {
	if e.StartedAt == nil {
		return 0
	}
	endTime := time.Now()
	if e.CompletedAt != nil {
		endTime = *e.CompletedAt
	}
	return endTime.Sub(*e.StartedAt)
}

// DurationMs returns execution duration in milliseconds
func (e *Execution) DurationMs() int64 {
	return e.Duration().Milliseconds()
}

// Progress returns execution progress as a percentage
func (e *Execution) Progress() float64 {
	if e.NodesTotal == 0 {
		return 0
	}
	return float64(e.NodesCompleted) / float64(e.NodesTotal) * 100
}
