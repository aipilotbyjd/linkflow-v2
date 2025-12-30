package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Permission represents granular RBAC permissions
type Permission struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"size:100;uniqueIndex;not null" json:"name"`
	Description *string   `gorm:"type:text" json:"description,omitempty"`
	Resource    string    `gorm:"size:50;not null;index" json:"resource"` // workflow, credential, etc.
	Action      string    `gorm:"size:30;not null" json:"action"`         // create, read, update, delete, execute
	CreatedAt   time.Time `json:"created_at"`
}

func (Permission) TableName() string {
	return "permissions"
}

// Role represents custom workspace roles
type Role struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID *uuid.UUID     `gorm:"type:uuid;index" json:"workspace_id,omitempty"` // nil = system role
	Name        string         `gorm:"size:50;not null" json:"name"`
	Description *string        `gorm:"type:text" json:"description,omitempty"`
	IsSystem    bool           `gorm:"default:false" json:"is_system"` // Cannot be deleted
	Color       *string        `gorm:"size:7" json:"color,omitempty"`  // Hex color
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Workspace   *Workspace       `gorm:"foreignKey:WorkspaceID" json:"-"`
	Permissions []RolePermission `gorm:"foreignKey:RoleID" json:"-"`
}

func (Role) TableName() string {
	return "roles"
}

// RolePermission links roles to permissions
type RolePermission struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	RoleID       uuid.UUID `gorm:"type:uuid;index;not null" json:"role_id"`
	PermissionID uuid.UUID `gorm:"type:uuid;index;not null" json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`

	Role       Role       `gorm:"foreignKey:RoleID" json:"-"`
	Permission Permission `gorm:"foreignKey:PermissionID" json:"-"`
}

func (RolePermission) TableName() string {
	return "role_permissions"
}

// DefaultPermissions returns the system-default permissions
func DefaultPermissions() []Permission {
	resources := []string{"workflow", "credential", "execution", "schedule", "webhook", "member", "settings", "billing"}
	actions := []string{"create", "read", "update", "delete", "execute"}

	var perms []Permission
	for _, resource := range resources {
		for _, action := range actions {
			// Skip non-applicable combinations
			if action == "execute" && resource != "workflow" {
				continue
			}
			perms = append(perms, Permission{
				Name:     resource + ":" + action,
				Resource: resource,
				Action:   action,
			})
		}
	}
	return perms
}

// DefaultRoles returns the system-default roles
func DefaultRoles() []Role {
	return []Role{
		{
			Name:        "owner",
			Description: strPtr("Full access to workspace"),
			IsSystem:    true,
		},
		{
			Name:        "admin",
			Description: strPtr("Administrative access"),
			IsSystem:    true,
		},
		{
			Name:        "editor",
			Description: strPtr("Can create and edit workflows"),
			IsSystem:    true,
		},
		{
			Name:        "viewer",
			Description: strPtr("Read-only access"),
			IsSystem:    true,
		},
		{
			Name:        "executor",
			Description: strPtr("Can only execute workflows"),
			IsSystem:    true,
		},
	}
}
