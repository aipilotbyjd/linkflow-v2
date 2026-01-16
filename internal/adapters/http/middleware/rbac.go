package middleware

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workspace"
)

type Permission string

const (
	PermissionRead   Permission = "read"
	PermissionWrite  Permission = "write"
	PermissionDelete Permission = "delete"
	PermissionAdmin  Permission = "admin"
)

var rolePermissions = map[workspace.Role][]Permission{
	workspace.RoleOwner:  {PermissionRead, PermissionWrite, PermissionDelete, PermissionAdmin},
	workspace.RoleAdmin:  {PermissionRead, PermissionWrite, PermissionDelete, PermissionAdmin},
	workspace.RoleMember: {PermissionRead, PermissionWrite},
	workspace.RoleViewer: {PermissionRead},
}

func RequireRBACPermission(permission Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wsCtx := GetWorkspaceFromContext(r.Context())
			if wsCtx == nil {
				common.Error(w, http.StatusForbidden, "FORBIDDEN", "workspace context required")
				return
			}

			permissions, ok := rolePermissions[wsCtx.Role]
			if !ok {
				common.Error(w, http.StatusForbidden, "FORBIDDEN", "invalid role")
				return
			}

			hasPermission := false
			for _, p := range permissions {
				if p == permission {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				common.Error(w, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireRBACRole(roles ...workspace.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wsCtx := GetWorkspaceFromContext(r.Context())
			if wsCtx == nil {
				common.Error(w, http.StatusForbidden, "FORBIDDEN", "workspace context required")
				return
			}

			hasRole := false
			for _, role := range roles {
				if wsCtx.Role == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				common.Error(w, http.StatusForbidden, "FORBIDDEN", "insufficient role")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
