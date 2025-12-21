package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Workflow struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID    uuid.UUID      `gorm:"type:uuid;index;not null" json:"workspace_id"`
	CreatedBy      uuid.UUID      `gorm:"type:uuid;not null" json:"created_by"`
	Name           string         `gorm:"size:255;not null" json:"name"`
	Description    *string        `gorm:"type:text" json:"description,omitempty"`
	Status         string         `gorm:"size:20;not null;default:draft;index" json:"status"`
	Version        int            `gorm:"not null;default:1" json:"version"`
	Nodes          JSONArray      `gorm:"type:jsonb;not null;default:'[]'" json:"nodes"`
	Connections    JSONArray      `gorm:"type:jsonb;not null;default:'[]'" json:"connections"`
	Settings       JSON           `gorm:"type:jsonb;default:'{}'" json:"settings"`
	Tags           StringArray    `gorm:"type:text[]" json:"tags"`
	FolderID       *uuid.UUID     `gorm:"type:uuid" json:"folder_id,omitempty"`
	ExecutionCount int            `gorm:"default:0" json:"execution_count"`
	LastExecutedAt *time.Time     `json:"last_executed_at,omitempty"`
	ActivatedAt    *time.Time     `json:"activated_at,omitempty"`
	ArchivedAt     *time.Time     `json:"archived_at,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Workspace  Workspace         `gorm:"foreignKey:WorkspaceID" json:"-"`
	Creator    User              `gorm:"foreignKey:CreatedBy" json:"-"`
	Versions   []WorkflowVersion `gorm:"foreignKey:WorkflowID" json:"-"`
	Executions []Execution       `gorm:"foreignKey:WorkflowID" json:"-"`
	Schedules  []Schedule        `gorm:"foreignKey:WorkflowID" json:"-"`
}

func (Workflow) TableName() string {
	return "workflows"
}

type WorkflowVersion struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkflowID    uuid.UUID  `gorm:"type:uuid;index;not null" json:"workflow_id"`
	Version       int        `gorm:"not null" json:"version"`
	Nodes         JSONArray  `gorm:"type:jsonb;not null" json:"nodes"`
	Connections   JSONArray  `gorm:"type:jsonb;not null" json:"connections"`
	Settings      JSON       `gorm:"type:jsonb" json:"settings"`
	CreatedBy     *uuid.UUID `gorm:"type:uuid" json:"created_by,omitempty"`
	ChangeMessage *string    `gorm:"type:text" json:"change_message,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`

	Workflow Workflow `gorm:"foreignKey:WorkflowID" json:"-"`
	Creator  *User    `gorm:"foreignKey:CreatedBy" json:"-"`
}

func (WorkflowVersion) TableName() string {
	return "workflow_versions"
}

type WorkflowFolder struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID uuid.UUID  `gorm:"type:uuid;index;not null" json:"workspace_id"`
	ParentID    *uuid.UUID `gorm:"type:uuid;index" json:"parent_id,omitempty"`
	Name        string     `gorm:"size:100;not null" json:"name"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	Workspace Workspace        `gorm:"foreignKey:WorkspaceID" json:"-"`
	Parent    *WorkflowFolder  `gorm:"foreignKey:ParentID" json:"-"`
	Children  []WorkflowFolder `gorm:"foreignKey:ParentID" json:"-"`
	Workflows []Workflow       `gorm:"foreignKey:FolderID" json:"-"`
}

func (WorkflowFolder) TableName() string {
	return "workflow_folders"
}
