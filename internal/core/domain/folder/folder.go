package folder

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Folder represents a folder for organizing workflows
type Folder struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID uuid.UUID      `gorm:"type:uuid;index;not null" json:"workspace_id"`
	ParentID    *uuid.UUID     `gorm:"type:uuid;index" json:"parent_id,omitempty"`
	Name        string         `gorm:"size:100;not null" json:"name"`
	Description *string        `gorm:"type:text" json:"description,omitempty"`
	Color       *string        `gorm:"size:20" json:"color,omitempty"`
	Position    int            `gorm:"default:0" json:"position"`
	CreatedBy   uuid.UUID      `gorm:"type:uuid;not null" json:"created_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Folder) TableName() string {
	return "folders"
}

// NewFolder creates a new folder
func NewFolder(workspaceID uuid.UUID, name string, createdBy uuid.UUID) (*Folder, error) {
	if workspaceID == uuid.Nil {
		return nil, ErrInvalidWorkspaceID
	}
	if name == "" {
		return nil, ErrNameRequired
	}
	if len(name) > 100 {
		return nil, ErrNameTooLong
	}
	if createdBy == uuid.Nil {
		return nil, ErrInvalidCreatedBy
	}

	return &Folder{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		Name:        name,
		Position:    0,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

// WithParent sets the parent folder
func (f *Folder) WithParent(parentID uuid.UUID) *Folder {
	f.ParentID = &parentID
	return f
}

// WithDescription sets the description
func (f *Folder) WithDescription(description string) *Folder {
	f.Description = &description
	return f
}

// WithColor sets the color
func (f *Folder) WithColor(color string) *Folder {
	f.Color = &color
	return f
}

// Update updates folder details
func (f *Folder) Update(name string, description, color *string) {
	f.Name = name
	f.Description = description
	f.Color = color
	f.UpdatedAt = time.Now()
}

// Move moves the folder to a new parent
func (f *Folder) Move(parentID *uuid.UUID) {
	f.ParentID = parentID
	f.UpdatedAt = time.Now()
}

// IsRoot checks if the folder is a root folder
func (f *Folder) IsRoot() bool {
	return f.ParentID == nil
}
