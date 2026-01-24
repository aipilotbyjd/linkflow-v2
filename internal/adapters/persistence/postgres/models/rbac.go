package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	WorkspaceID *uuid.UUID `gorm:"type:uuid;index"` // Null for system roles
	Name        string     `gorm:"size:100;not null"`
	Description string     `gorm:"type:text"`
	IsProtected bool       `gorm:"default:false"`
	IsDefault   bool       `gorm:"default:false"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	Permissions []Permission `gorm:"many2many:role_permissions;"`
}

func (Role) TableName() string {
	return "roles"
}

type Permission struct {
	ID          string `gorm:"size:100;primaryKey"` // e.g. "workflow:create"
	Scope       string `gorm:"size:50;not null;index"`
	Name        string `gorm:"size:100;not null"`
	Description string `gorm:"type:text"`
	CreatedAt   time.Time
}

func (Permission) TableName() string {
	return "permissions"
}

// Join table (implicit in GORM but explicit if we need extra fields, here standard)
// We let GORM handle role_permissions table via many2many
