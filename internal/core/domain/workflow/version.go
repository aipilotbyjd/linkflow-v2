package workflow

import (
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// Version represents a workflow version snapshot
type Version struct {
	ID            uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkflowID    uuid.UUID       `gorm:"type:uuid;index;not null" json:"workflow_id"`
	Version       int             `gorm:"not null" json:"version"`
	Nodes         types.JSONArray `gorm:"type:jsonb;not null" json:"nodes"`
	Connections   types.JSONArray `gorm:"type:jsonb;not null" json:"connections"`
	Settings      types.JSON      `gorm:"type:jsonb" json:"settings"`
	CreatedBy     *uuid.UUID      `gorm:"type:uuid" json:"created_by,omitempty"`
	ChangeMessage *string         `gorm:"type:text" json:"change_message,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`

	Workflow Workflow `gorm:"foreignKey:WorkflowID" json:"-"`
}

func (Version) TableName() string {
	return "workflow_versions"
}

// NewVersion creates a new version from a workflow
func NewVersion(workflow *Workflow, changeMessage *string) *Version {
	return &Version{
		ID:            uuid.New(),
		WorkflowID:    workflow.ID,
		Version:       workflow.Version,
		Nodes:         workflow.Nodes,
		Connections:   workflow.Connections,
		Settings:      workflow.Settings,
		CreatedBy:     &workflow.CreatedBy,
		ChangeMessage: changeMessage,
		CreatedAt:     time.Now(),
	}
}

// Folder represents a workflow folder/project
type Folder struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID uuid.UUID  `gorm:"type:uuid;index;not null" json:"workspace_id"`
	ParentID    *uuid.UUID `gorm:"type:uuid;index" json:"parent_id,omitempty"`
	Name        string     `gorm:"size:100;not null" json:"name"`
	Description *string    `gorm:"type:text" json:"description,omitempty"`
	Color       *string    `gorm:"size:20" json:"color,omitempty"`
	Icon        *string    `gorm:"size:50" json:"icon,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (Folder) TableName() string {
	return "projects"
}

// NewFolder creates a new folder
func NewFolder(workspaceID uuid.UUID, name string, parentID *uuid.UUID) *Folder {
	return &Folder{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		ParentID:    parentID,
		Name:        name,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// Update updates folder details
func (f *Folder) Update(name string, description, color, icon *string) {
	f.Name = name
	f.Description = description
	f.Color = color
	f.Icon = icon
	f.UpdatedAt = time.Now()
}

// Move moves folder to new parent
func (f *Folder) Move(parentID *uuid.UUID) {
	f.ParentID = parentID
	f.UpdatedAt = time.Now()
}
