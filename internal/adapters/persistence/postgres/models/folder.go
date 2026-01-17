package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Folder struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	WorkspaceID uuid.UUID  `gorm:"type:uuid;index;not null"`
	ParentID    *uuid.UUID `gorm:"type:uuid;index"`
	Name        string     `gorm:"size:100;not null"`
	Description *string    `gorm:"type:text"`
	Color       *string    `gorm:"size:20"`
	Position    int        `gorm:"default:0"`
	CreatedBy   uuid.UUID  `gorm:"type:uuid;not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	Workspace Workspace `gorm:"foreignKey:WorkspaceID"`
	Parent    *Folder   `gorm:"foreignKey:ParentID"`
	Creator   User      `gorm:"foreignKey:CreatedBy"`
}

func (Folder) TableName() string {
	return "folders"
}
