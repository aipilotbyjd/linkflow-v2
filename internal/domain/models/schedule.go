package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Schedule struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkflowID      uuid.UUID      `gorm:"type:uuid;index;not null" json:"workflow_id"`
	WorkspaceID     uuid.UUID      `gorm:"type:uuid;index;not null" json:"workspace_id"`
	CreatedBy       uuid.UUID      `gorm:"type:uuid;not null" json:"created_by"`
	Name            string         `gorm:"size:100;not null" json:"name"`
	Description     *string        `gorm:"type:text" json:"description,omitempty"`
	CronExpression  string         `gorm:"size:100;not null" json:"cron_expression"`
	Timezone        string         `gorm:"size:50;default:UTC" json:"timezone"`
	IsActive        bool           `gorm:"default:true" json:"is_active"`
	InputData       JSON           `gorm:"type:jsonb" json:"input_data,omitempty"`
	NextRunAt       *time.Time     `gorm:"index" json:"next_run_at,omitempty"`
	LastRunAt       *time.Time     `json:"last_run_at,omitempty"`
	LastExecutionID *uuid.UUID     `gorm:"type:uuid" json:"last_execution_id,omitempty"`
	RunCount        int            `gorm:"default:0" json:"run_count"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	Workflow      Workflow   `gorm:"foreignKey:WorkflowID" json:"-"`
	Workspace     Workspace  `gorm:"foreignKey:WorkspaceID" json:"-"`
	Creator       User       `gorm:"foreignKey:CreatedBy" json:"-"`
	LastExecution *Execution `gorm:"foreignKey:LastExecutionID" json:"-"`
}

func (Schedule) TableName() string {
	return "schedules"
}

// GetWorkspaceID implements the WorkspaceOwned interface for authorization checks
func (s *Schedule) GetWorkspaceID() uuid.UUID {
	return s.WorkspaceID
}
