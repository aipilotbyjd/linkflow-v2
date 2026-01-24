package mappers

import (
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/models"
	"github.com/linkflow-ai/linkflow/internal/core/domain/rbac"
)

func ToDomainRole(m *models.Role) *rbac.Role {
	if m == nil {
		return nil
	}
	role := &rbac.Role{
		ID:          m.ID,
		WorkspaceID: m.WorkspaceID,
		Name:        m.Name,
		Description: m.Description,
		IsProtected: m.IsProtected,
		IsDefault:   m.IsDefault,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		Permissions: make([]rbac.Permission, len(m.Permissions)),
	}

	for i, p := range m.Permissions {
		role.Permissions[i] = ToDomainPermission(&p)
	}

	return role
}

func ToDomainPermission(m *models.Permission) rbac.Permission {
	if m == nil {
		return rbac.Permission{}
	}
	return rbac.Permission{
		ID:          m.ID,
		Scope:       m.Scope,
		Name:        m.Name,
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
	}
}

func ToModelRole(d *rbac.Role) *models.Role {
	if d == nil {
		return nil
	}
	return &models.Role{
		ID:          d.ID,
		WorkspaceID: d.WorkspaceID,
		Name:        d.Name,
		Description: d.Description,
		IsProtected: d.IsProtected,
		IsDefault:   d.IsDefault,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

func ToModelPermission(d rbac.Permission) models.Permission {
	return models.Permission{
		ID:          d.ID,
		Scope:       d.Scope,
		Name:        d.Name,
		Description: d.Description,
		CreatedAt:   d.CreatedAt,
	}
}
