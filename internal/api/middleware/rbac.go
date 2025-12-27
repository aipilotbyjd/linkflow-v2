package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"gorm.io/gorm"
)

type RBACMiddleware struct {
	db *gorm.DB
}

func NewRBACMiddleware(db *gorm.DB) *RBACMiddleware {
	return &RBACMiddleware{db: db}
}

// Permission constants
const (
	PermWorkflowCreate  = "workflow:create"
	PermWorkflowRead    = "workflow:read"
	PermWorkflowUpdate  = "workflow:update"
	PermWorkflowDelete  = "workflow:delete"
	PermWorkflowExecute = "workflow:execute"

	PermCredentialCreate = "credential:create"
	PermCredentialRead   = "credential:read"
	PermCredentialUpdate = "credential:update"
	PermCredentialDelete = "credential:delete"

	PermExecutionRead   = "execution:read"
	PermExecutionDelete = "execution:delete"

	PermScheduleCreate = "schedule:create"
	PermScheduleRead   = "schedule:read"
	PermScheduleUpdate = "schedule:update"
	PermScheduleDelete = "schedule:delete"

	PermMemberRead   = "member:read"
	PermMemberCreate = "member:create"
	PermMemberUpdate = "member:update"
	PermMemberDelete = "member:delete"

	PermSettingsRead   = "settings:read"
	PermSettingsUpdate = "settings:update"

	PermBillingRead   = "billing:read"
	PermBillingUpdate = "billing:update"
)

// Role permission mappings
var rolePermissions = map[string][]string{
	"owner": {
		PermWorkflowCreate, PermWorkflowRead, PermWorkflowUpdate, PermWorkflowDelete, PermWorkflowExecute,
		PermCredentialCreate, PermCredentialRead, PermCredentialUpdate, PermCredentialDelete,
		PermExecutionRead, PermExecutionDelete,
		PermScheduleCreate, PermScheduleRead, PermScheduleUpdate, PermScheduleDelete,
		PermMemberRead, PermMemberCreate, PermMemberUpdate, PermMemberDelete,
		PermSettingsRead, PermSettingsUpdate,
		PermBillingRead, PermBillingUpdate,
	},
	"admin": {
		PermWorkflowCreate, PermWorkflowRead, PermWorkflowUpdate, PermWorkflowDelete, PermWorkflowExecute,
		PermCredentialCreate, PermCredentialRead, PermCredentialUpdate, PermCredentialDelete,
		PermExecutionRead, PermExecutionDelete,
		PermScheduleCreate, PermScheduleRead, PermScheduleUpdate, PermScheduleDelete,
		PermMemberRead, PermMemberCreate, PermMemberUpdate,
		PermSettingsRead, PermSettingsUpdate,
		PermBillingRead,
	},
	"editor": {
		PermWorkflowCreate, PermWorkflowRead, PermWorkflowUpdate, PermWorkflowExecute,
		PermCredentialCreate, PermCredentialRead, PermCredentialUpdate,
		PermExecutionRead,
		PermScheduleCreate, PermScheduleRead, PermScheduleUpdate,
		PermMemberRead,
		PermSettingsRead,
	},
	"viewer": {
		PermWorkflowRead,
		PermCredentialRead,
		PermExecutionRead,
		PermScheduleRead,
		PermMemberRead,
		PermSettingsRead,
	},
	"executor": {
		PermWorkflowRead, PermWorkflowExecute,
		PermExecutionRead,
	},
}

// RequirePermission returns middleware that checks for a specific permission
func (m *RBACMiddleware) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wsCtx := GetWorkspaceFromContext(r.Context())
			if wsCtx == nil {
				dto.ErrorResponse(w, http.StatusForbidden, "workspace context required")
				return
			}

			if !m.hasPermission(wsCtx.Role, permission) {
				dto.ErrorResponse(w, http.StatusForbidden, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermission returns middleware that checks for any of the specified permissions
func (m *RBACMiddleware) RequireAnyPermission(permissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wsCtx := GetWorkspaceFromContext(r.Context())
			if wsCtx == nil {
				dto.ErrorResponse(w, http.StatusForbidden, "workspace context required")
				return
			}

			for _, perm := range permissions {
				if m.hasPermission(wsCtx.Role, perm) {
					next.ServeHTTP(w, r)
					return
				}
			}

			dto.ErrorResponse(w, http.StatusForbidden, "insufficient permissions")
		})
	}
}

// RequireAllPermissions returns middleware that checks for all specified permissions
func (m *RBACMiddleware) RequireAllPermissions(permissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wsCtx := GetWorkspaceFromContext(r.Context())
			if wsCtx == nil {
				dto.ErrorResponse(w, http.StatusForbidden, "workspace context required")
				return
			}

			for _, perm := range permissions {
				if !m.hasPermission(wsCtx.Role, perm) {
					dto.ErrorResponse(w, http.StatusForbidden, "insufficient permissions")
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole returns middleware that checks for a specific role or higher
func (m *RBACMiddleware) RequireRole(minRole string) func(http.Handler) http.Handler {
	roleHierarchy := map[string]int{
		"owner":    5,
		"admin":    4,
		"editor":   3,
		"viewer":   2,
		"executor": 1,
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wsCtx := GetWorkspaceFromContext(r.Context())
			if wsCtx == nil {
				dto.ErrorResponse(w, http.StatusForbidden, "workspace context required")
				return
			}

			userLevel, ok := roleHierarchy[wsCtx.Role]
			if !ok {
				dto.ErrorResponse(w, http.StatusForbidden, "invalid role")
				return
			}

			requiredLevel, ok := roleHierarchy[minRole]
			if !ok {
				dto.ErrorResponse(w, http.StatusInternalServerError, "invalid required role")
				return
			}

			if userLevel < requiredLevel {
				dto.ErrorResponse(w, http.StatusForbidden, "insufficient role level")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireOwner returns middleware that only allows workspace owners
func (m *RBACMiddleware) RequireOwner() func(http.Handler) http.Handler {
	return m.RequireRole("owner")
}

// RequireAdmin returns middleware that allows admins and above
func (m *RBACMiddleware) RequireAdmin() func(http.Handler) http.Handler {
	return m.RequireRole("admin")
}

// RequireEditor returns middleware that allows editors and above
func (m *RBACMiddleware) RequireEditor() func(http.Handler) http.Handler {
	return m.RequireRole("editor")
}

func (m *RBACMiddleware) hasPermission(role, permission string) bool {
	perms, ok := rolePermissions[role]
	if !ok {
		return false
	}

	for _, p := range perms {
		if p == permission {
			return true
		}
	}
	return false
}

// CheckPermission checks if user has permission (for use in handlers)
func (m *RBACMiddleware) CheckPermission(ctx context.Context, permission string) bool {
	wsCtx := GetWorkspaceFromContext(ctx)
	if wsCtx == nil {
		return false
	}
	return m.hasPermission(wsCtx.Role, permission)
}

// GetUserPermissions returns all permissions for a role
func (m *RBACMiddleware) GetUserPermissions(role string) []string {
	if perms, ok := rolePermissions[role]; ok {
		return perms
	}
	return []string{}
}

// CheckFeatureAccess checks if the workspace plan allows the feature
func (m *RBACMiddleware) CheckFeatureAccess(ctx context.Context, workspaceID uuid.UUID, feature string) bool {
	var workspace models.Workspace
	if err := m.db.First(&workspace, "id = ?", workspaceID).Error; err != nil {
		return false
	}

	var plan models.Plan
	if err := m.db.First(&plan, "id = ?", workspace.PlanID).Error; err != nil {
		return false
	}

	// Check feature in plan features
	var features models.PlanFeatures
	if plan.Features != nil {
		featuresBytes, _ := plan.Features.Value()
		if bytes, ok := featuresBytes.([]byte); ok {
			if err := json.Unmarshal(bytes, &features); err != nil {
				return false
			}
		}
	}

	switch feature {
	case "audit_logs":
		return features.AuditLogs
	case "workflow_comments":
		return features.WorkflowComments
	case "team_roles":
		return features.TeamRoles
	case "api_access":
		return features.APIAccess
	case "webhooks":
		return features.Webhooks
	case "schedules":
		return features.Schedules
	default:
		return true // Allow unknown features by default
	}
}

// RequireFeature returns middleware that checks for a plan feature
func (m *RBACMiddleware) RequireFeature(feature string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wsCtx := GetWorkspaceFromContext(r.Context())
			if wsCtx == nil {
				dto.ErrorResponse(w, http.StatusForbidden, "workspace context required")
				return
			}

			if !m.CheckFeatureAccess(r.Context(), wsCtx.WorkspaceID, feature) {
				dto.ErrorResponse(w, http.StatusForbidden, "feature not available on your plan")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
