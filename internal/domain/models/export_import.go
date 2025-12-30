package models

import (
	"time"

	"github.com/google/uuid"
)

// WorkflowExport tracks workflow exports
type WorkflowExport struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkflowID         uuid.UUID `gorm:"type:uuid;index;not null" json:"workflow_id"`
	WorkspaceID        uuid.UUID `gorm:"type:uuid;index;not null" json:"workspace_id"`
	ExportedBy         uuid.UUID `gorm:"type:uuid;not null" json:"exported_by"`
	Version            int       `gorm:"not null" json:"version"`
	Format             string    `gorm:"size:20;not null;default:json" json:"format"` // json, yaml
	FileSize           int       `gorm:"default:0" json:"file_size"`
	IncludeCredentials bool      `gorm:"default:false" json:"include_credentials"`
	CreatedAt          time.Time `json:"created_at"`

	Workflow  Workflow  `gorm:"foreignKey:WorkflowID" json:"-"`
	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
	Exporter  User      `gorm:"foreignKey:ExportedBy" json:"-"`
}

func (WorkflowExport) TableName() string {
	return "workflow_exports"
}

// WorkflowImport tracks workflow imports
type WorkflowImport struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkflowID  *uuid.UUID `gorm:"type:uuid;index" json:"workflow_id,omitempty"` // Created workflow
	WorkspaceID uuid.UUID  `gorm:"type:uuid;index;not null" json:"workspace_id"`
	ImportedBy  uuid.UUID  `gorm:"type:uuid;not null" json:"imported_by"`
	SourceName  *string    `gorm:"size:255" json:"source_name,omitempty"` // Original workflow name
	SourceType  string     `gorm:"size:20;not null" json:"source_type"`   // file, url, template
	Status      string     `gorm:"size:20;not null;default:pending" json:"status"`
	Error       *string    `gorm:"type:text" json:"error,omitempty"`
	Warnings    JSON       `gorm:"type:jsonb" json:"warnings,omitempty"` // Non-fatal issues
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	Workflow  *Workflow `gorm:"foreignKey:WorkflowID" json:"-"`
	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
	Importer  User      `gorm:"foreignKey:ImportedBy" json:"-"`
}

func (WorkflowImport) TableName() string {
	return "workflow_imports"
}
