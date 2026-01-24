package rbac

import (
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/rbac"
)

type RoleResponse struct {
	ID          uuid.UUID            `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	IsProtected bool                 `json:"is_protected"`
	IsDefault   bool                 `json:"is_default"`
	Permissions []PermissionResponse `json:"permissions"`
}

type PermissionResponse struct {
	ID          string `json:"id"`
	Scope       string `json:"scope"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateRoleRequest struct {
	Name        string   `json:"name" validate:"required,min=3,max=50"`
	Description string   `json:"description" validate:"max=200"`
	Permissions []string `json:"permissions" validate:"required"` // List of permission IDs
}

type UpdateRoleRequest struct {
	Name        string   `json:"name" validate:"omitempty,min=3,max=50"`
	Description string   `json:"description" validate:"max=200"`
	Permissions []string `json:"permissions" validate:"omitempty"`
}

func ToRoleResponse(r *rbac.Role) RoleResponse {
	perms := make([]PermissionResponse, len(r.Permissions))
	for i, p := range r.Permissions {
		perms[i] = PermissionResponse{
			ID:          p.ID,
			Scope:       p.Scope,
			Name:        p.Name,
			Description: p.Description,
		}
	}
	return RoleResponse{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		IsProtected: r.IsProtected,
		IsDefault:   r.IsDefault,
		Permissions: perms,
	}
}

func ToRoleResponseList(roles []rbac.Role) []RoleResponse {
	res := make([]RoleResponse, len(roles))
	for i, r := range roles {
		res[i] = ToRoleResponse(&r)
	}
	return res
}
