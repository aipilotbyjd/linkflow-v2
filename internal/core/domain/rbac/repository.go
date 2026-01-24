package rbac

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	// Role Management
	CreateRole(ctx context.Context, role *Role) error
	UpdateRole(ctx context.Context, role *Role) error
	DeleteRole(ctx context.Context, id uuid.UUID) error
	GetRole(ctx context.Context, id uuid.UUID) (*Role, error)
	GetRoleByName(ctx context.Context, workspaceID *uuid.UUID, name string) (*Role, error)

	// List roles (system defaults + workspace specific)
	ListRoles(ctx context.Context, workspaceID uuid.UUID) ([]Role, error)
	ListSystemRoles(ctx context.Context) ([]Role, error)

	// Permission Management
	ListPermissions(ctx context.Context) ([]Permission, error)
	GetPermissionsForRole(ctx context.Context, roleID uuid.UUID) ([]Permission, error)
	AssignPermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []string) error
	RemovePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []string) error
}
