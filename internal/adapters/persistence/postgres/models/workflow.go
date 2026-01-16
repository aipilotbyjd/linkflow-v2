package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
	"gorm.io/gorm"
)

type Workflow struct {
	ID             uuid.UUID         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	WorkspaceID    uuid.UUID         `gorm:"type:uuid;index;not null"`
	CreatedBy      uuid.UUID         `gorm:"type:uuid;not null"`
	Name           string            `gorm:"size:255;not null"`
	Description    *string           `gorm:"type:text"`
	Status         string            `gorm:"size:20;not null;default:draft;index"`
	Version        int               `gorm:"default:1"`
	Nodes          types.JSONArray   `gorm:"type:jsonb;not null;default:'[]'"`
	Connections    types.JSONArray   `gorm:"type:jsonb;not null;default:'[]'"`
	Settings       types.JSON        `gorm:"type:jsonb;default:'{}'"`
	Tags           types.StringArray `gorm:"type:text[]"`
	ExecutionCount int               `gorm:"default:0"`
	LastExecutedAt *time.Time
	ActivatedAt    *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`

	Workspace Workspace         `gorm:"foreignKey:WorkspaceID"`
	Creator   User              `gorm:"foreignKey:CreatedBy"`
	Versions  []WorkflowVersion `gorm:"foreignKey:WorkflowID"`
}

func (Workflow) TableName() string {
	return "workflows"
}

type WorkflowVersion struct {
	ID          uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	WorkflowID  uuid.UUID       `gorm:"type:uuid;index;not null"`
	Version     int             `gorm:"not null"`
	Nodes       types.JSONArray `gorm:"type:jsonb;not null"`
	Connections types.JSONArray `gorm:"type:jsonb;not null"`
	Settings    types.JSON      `gorm:"type:jsonb"`
	ChangeNote  *string         `gorm:"type:text"`
	CreatedBy   uuid.UUID       `gorm:"type:uuid;not null"`
	CreatedAt   time.Time

	Workflow Workflow `gorm:"foreignKey:WorkflowID"`
	Creator  User     `gorm:"foreignKey:CreatedBy"`
}

func (WorkflowVersion) TableName() string {
	return "workflow_versions"
}
