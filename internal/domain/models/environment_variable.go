package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EnvironmentVariable represents workspace-level secrets/variables
type EnvironmentVariable struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID uuid.UUID      `gorm:"type:uuid;index;not null" json:"workspace_id"`
	CreatedBy   uuid.UUID      `gorm:"type:uuid;not null" json:"created_by"`
	Name        string         `gorm:"size:100;not null" json:"name"`
	Value       string         `gorm:"type:text;not null" json:"-"`        // Encrypted
	IsSecret    bool           `gorm:"default:false" json:"is_secret"`     // If true, mask in logs
	Environment *string        `gorm:"size:20" json:"environment,omitempty"` // dev, staging, prod
	Description *string        `gorm:"type:text" json:"description,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
	Creator   User      `gorm:"foreignKey:CreatedBy" json:"-"`
}

func (EnvironmentVariable) TableName() string {
	return "environment_variables"
}
