package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WorkflowComment represents annotations on workflows/nodes
type WorkflowComment struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkflowID  uuid.UUID      `gorm:"type:uuid;index;not null" json:"workflow_id"`
	WorkspaceID uuid.UUID      `gorm:"type:uuid;index;not null" json:"workspace_id"`
	NodeID      *string        `gorm:"size:100;index" json:"node_id,omitempty"`    // nil = workflow-level comment
	ParentID    *uuid.UUID     `gorm:"type:uuid;index" json:"parent_id,omitempty"` // For replies
	CreatedBy   uuid.UUID      `gorm:"type:uuid;not null" json:"created_by"`
	Content     string         `gorm:"type:text;not null" json:"content"`
	IsResolved  bool           `gorm:"default:false" json:"is_resolved"`
	ResolvedBy  *uuid.UUID     `gorm:"type:uuid" json:"resolved_by,omitempty"`
	ResolvedAt  *time.Time     `json:"resolved_at,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Workflow  Workflow          `gorm:"foreignKey:WorkflowID" json:"-"`
	Workspace Workspace         `gorm:"foreignKey:WorkspaceID" json:"-"`
	Creator   User              `gorm:"foreignKey:CreatedBy" json:"-"`
	Resolver  *User             `gorm:"foreignKey:ResolvedBy" json:"-"`
	Parent    *WorkflowComment  `gorm:"foreignKey:ParentID" json:"-"`
	Replies   []WorkflowComment `gorm:"foreignKey:ParentID" json:"-"`
}

func (WorkflowComment) TableName() string {
	return "workflow_comments"
}
