package models

import (
	"time"

	"github.com/google/uuid"
)

// ExecutionShare represents shareable execution links
type ExecutionShare struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ExecutionID   uuid.UUID  `gorm:"type:uuid;index;not null" json:"execution_id"`
	WorkspaceID   uuid.UUID  `gorm:"type:uuid;index;not null" json:"workspace_id"`
	CreatedBy     uuid.UUID  `gorm:"type:uuid;not null" json:"created_by"`
	Token         string     `gorm:"size:64;uniqueIndex;not null" json:"token"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	Password      *string    `gorm:"size:255" json:"-"` // Optional password protection
	ViewCount     int        `gorm:"default:0" json:"view_count"`
	MaxViews      *int       `json:"max_views,omitempty"` // Limit views
	AllowDownload bool       `gorm:"default:false" json:"allow_download"`
	IncludeLogs   bool       `gorm:"default:true" json:"include_logs"`
	IncludeData   bool       `gorm:"default:false" json:"include_data"` // Include input/output data
	CreatedAt     time.Time  `json:"created_at"`

	Execution Execution `gorm:"foreignKey:ExecutionID" json:"-"`
	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
	Creator   User      `gorm:"foreignKey:CreatedBy" json:"-"`
}

func (ExecutionShare) TableName() string {
	return "execution_shares"
}
