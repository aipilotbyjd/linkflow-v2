package rbac

import (
	"time"

	"github.com/google/uuid"
)

// Permission represents a granular access control permission
type Permission struct {
	ID          string    `json:"id"`          // e.g., "workflow:create"
	Scope       string    `json:"scope"`       // e.g., "workflow"
	Name        string    `json:"name"`        // e.g., "Create Workflows"
	Description string    `json:"description"` // e.g., "Allows creating new workflows"
	CreatedAt   time.Time `json:"created_at"`
}

// Role represents a collection of permissions
type Role struct {
	ID          uuid.UUID    `json:"id"`
	WorkspaceID *uuid.UUID   `json:"workspace_id,omitempty"` // Null for system roles
	Name        string       `json:"name"`
	Description string       `json:"description"`
	IsProtected bool         `json:"is_protected"` // e.g., Owner role cannot be deleted
	IsDefault   bool         `json:"is_default"`   // e.g., Default role for new members
	Permissions []Permission `json:"permissions,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// NewRole creates a new role instance
func NewRole(workspaceID *uuid.UUID, name string, description string) *Role {
	now := time.Now()
	return &Role{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		Name:        name,
		Description: description,
		Permissions: []Permission{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// IsSystem returns true if the role is a system role (no workspace ID)
func (r *Role) IsSystem() bool {
	return r.WorkspaceID == nil
}

// HasPermission checks if the role has a specific permission
func (r *Role) HasPermission(permID string) bool {
	for _, p := range r.Permissions {
		if p.ID == permID {
			return true
		}
	}
	return false
}
