package models

import (
	"time"

	"github.com/google/uuid"
)

type Execution struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkflowID        uuid.UUID  `gorm:"type:uuid;index;not null" json:"workflow_id"`
	WorkspaceID       uuid.UUID  `gorm:"type:uuid;index;not null" json:"workspace_id"`
	TriggeredBy       *uuid.UUID `gorm:"type:uuid" json:"triggered_by,omitempty"`
	WorkflowVersion   int        `gorm:"not null" json:"workflow_version"`
	Status            string     `gorm:"size:20;not null;default:queued;index" json:"status"`
	TriggerType       string     `gorm:"size:20;not null" json:"trigger_type"`
	TriggerData       JSON       `gorm:"type:jsonb" json:"trigger_data,omitempty"`
	InputData         JSON       `gorm:"type:jsonb" json:"input_data,omitempty"`
	OutputData        JSON       `gorm:"type:jsonb" json:"output_data,omitempty"`
	ErrorMessage      *string    `gorm:"type:text" json:"error_message,omitempty"`
	ErrorNodeID       *string    `gorm:"size:100" json:"error_node_id,omitempty"`
	QueuedAt          time.Time  `gorm:"default:now()" json:"queued_at"`
	StartedAt         *time.Time `json:"started_at,omitempty"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
	NodesTotal        int        `gorm:"default:0" json:"nodes_total"`
	NodesCompleted    int        `gorm:"default:0" json:"nodes_completed"`
	RetryCount        int        `gorm:"default:0" json:"retry_count"`
	ParentExecutionID *uuid.UUID `gorm:"type:uuid" json:"parent_execution_id,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`

	// Relations
	Workflow        Workflow        `gorm:"foreignKey:WorkflowID" json:"-"`
	Workspace       Workspace       `gorm:"foreignKey:WorkspaceID" json:"-"`
	Trigger         *User           `gorm:"foreignKey:TriggeredBy" json:"-"`
	ParentExecution *Execution      `gorm:"foreignKey:ParentExecutionID" json:"-"`
	NodeExecutions  []NodeExecution `gorm:"foreignKey:ExecutionID" json:"-"`
}

func (Execution) TableName() string {
	return "executions"
}

type NodeExecution struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ExecutionID  uuid.UUID  `gorm:"type:uuid;index;not null" json:"execution_id"`
	NodeID       string     `gorm:"size:100;not null" json:"node_id"`
	NodeType     string     `gorm:"size:50;not null" json:"node_type"`
	NodeName     *string    `gorm:"size:255" json:"node_name,omitempty"`
	Status       string     `gorm:"size:20;not null;default:pending;index" json:"status"`
	InputData    JSON       `gorm:"type:jsonb" json:"input_data,omitempty"`
	OutputData   JSON       `gorm:"type:jsonb" json:"output_data,omitempty"`
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

type ExecutionLog struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ExecutionID uuid.UUID `gorm:"type:uuid;index;not null" json:"execution_id"`
	NodeID      *string   `gorm:"size:100" json:"node_id,omitempty"`
	Level       string    `gorm:"size:10;not null" json:"level"`
	Message     string    `gorm:"type:text;not null" json:"message"`
	Data        JSON      `gorm:"type:jsonb" json:"data,omitempty"`
	Timestamp   time.Time `gorm:"not null;index" json:"timestamp"`

	Execution Execution `gorm:"foreignKey:ExecutionID" json:"-"`
}

func (ExecutionLog) TableName() string {
	return "execution_logs"
}
