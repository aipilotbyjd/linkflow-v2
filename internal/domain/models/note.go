package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Note represents a sticky note attached to any resource (workflow, execution, etc.)
type Note struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID  uuid.UUID      `gorm:"type:uuid;index;not null" json:"workspace_id"`
	ResourceID   uuid.UUID      `gorm:"type:uuid;index;not null" json:"resource_id"`
	ResourceName string         `gorm:"size:50;index;not null" json:"resource_name"` // workflow, execution, etc.
	CreatedBy    uuid.UUID      `gorm:"type:uuid;not null" json:"created_by"`
	Content      string         `gorm:"type:text;not null" json:"content"`
	Position     JSON           `gorm:"type:jsonb;default:'{}'" json:"position"` // {x, y}
	Size         JSON           `gorm:"type:jsonb;default:'{}'" json:"size"`     // {width, height}
	Color        string         `gorm:"size:20;default:'yellow'" json:"color"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
	Creator   User      `gorm:"foreignKey:CreatedBy" json:"-"`
}

func (Note) TableName() string {
	return "notes"
}

// GetWorkspaceID implements the WorkspaceOwned interface
func (n *Note) GetWorkspaceID() uuid.UUID {
	return n.WorkspaceID
}

// Note resource name constants
const (
	NoteResourceWorkflow  = "workflow"
	NoteResourceExecution = "execution"
)
