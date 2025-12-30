package models

import (
	"time"

	"github.com/google/uuid"
)

// SubWorkflowExecution tracks sub-workflow executions
type SubWorkflowExecution struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ParentExecutionID uuid.UUID  `gorm:"type:uuid;index;not null" json:"parent_execution_id"`
	ChildExecutionID  uuid.UUID  `gorm:"type:uuid;index;not null" json:"child_execution_id"`
	ParentNodeID      string     `gorm:"size:100;not null" json:"parent_node_id"`
	InputMapping      JSON       `gorm:"type:jsonb" json:"input_mapping,omitempty"`
	OutputMapping     JSON       `gorm:"type:jsonb" json:"output_mapping,omitempty"`
	Status            string     `gorm:"size:20;not null;default:pending" json:"status"`
	CreatedAt         time.Time  `json:"created_at"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`

	ParentExecution Execution `gorm:"foreignKey:ParentExecutionID" json:"-"`
	ChildExecution  Execution `gorm:"foreignKey:ChildExecutionID" json:"-"`
}

func (SubWorkflowExecution) TableName() string {
	return "sub_workflow_executions"
}
