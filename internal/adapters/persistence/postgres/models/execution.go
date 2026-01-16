package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type Execution struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	WorkflowID        uuid.UUID  `gorm:"type:uuid;index;not null"`
	WorkspaceID       uuid.UUID  `gorm:"type:uuid;index;not null"`
	TriggeredBy       *uuid.UUID `gorm:"type:uuid"`
	WorkflowVersion   int        `gorm:"not null"`
	Status            string     `gorm:"size:20;not null;default:queued;index"`
	TriggerType       string     `gorm:"size:20;not null"`
	TriggerData       types.JSON `gorm:"type:jsonb"`
	InputData         types.JSON `gorm:"type:jsonb"`
	OutputData        types.JSON `gorm:"type:jsonb"`
	ErrorMessage      *string    `gorm:"type:text"`
	ErrorNodeID       *string    `gorm:"size:100"`
	QueuedAt          time.Time  `gorm:"default:now()"`
	StartedAt         *time.Time
	CompletedAt       *time.Time
	PausedAt          *time.Time
	ResumedAt         *time.Time
	NodesTotal        int        `gorm:"default:0"`
	NodesCompleted    int        `gorm:"default:0"`
	RetryCount        int        `gorm:"default:0"`
	MaxRetries        int        `gorm:"default:3"`
	Priority          int        `gorm:"default:5;index"`
	TimeoutSeconds    int        `gorm:"default:3600"`
	ParentExecutionID *uuid.UUID `gorm:"type:uuid"`
	BatchID           *uuid.UUID `gorm:"type:uuid;index"`
	CreatedAt         time.Time

	Workflow       Workflow        `gorm:"foreignKey:WorkflowID"`
	Workspace      Workspace       `gorm:"foreignKey:WorkspaceID"`
	NodeExecutions []NodeExecution `gorm:"foreignKey:ExecutionID"`
}

func (Execution) TableName() string {
	return "executions"
}

type NodeExecution struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ExecutionID  uuid.UUID  `gorm:"type:uuid;index;not null"`
	NodeID       string     `gorm:"size:100;not null"`
	NodeType     string     `gorm:"size:50;not null"`
	NodeName     *string    `gorm:"size:255"`
	Status       string     `gorm:"size:20;not null;default:pending;index"`
	InputData    types.JSON `gorm:"type:jsonb"`
	OutputData   types.JSON `gorm:"type:jsonb"`
	ErrorMessage *string    `gorm:"type:text"`
	StartedAt    *time.Time
	CompletedAt  *time.Time
	DurationMs   *int
	RetryCount   int `gorm:"default:0"`
	CreatedAt    time.Time

	Execution Execution `gorm:"foreignKey:ExecutionID"`
}

func (NodeExecution) TableName() string {
	return "node_executions"
}
