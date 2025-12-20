package middleware

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
)

type TenantMiddleware struct {
	workspaceSvc *services.WorkspaceService
}

func NewTenantMiddleware(workspaceSvc *services.WorkspaceService) *TenantMiddleware {
	return &TenantMiddleware{workspaceSvc: workspaceSvc}
}

type WorkspaceContext struct {
	WorkspaceID uuid.UUID
	Role        string
}

func (m *TenantMiddleware) RequireMembership(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := GetUserFromContext(r.Context())
		if claims == nil {
			dto.ErrorResponse(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		workspaceIDStr := chi.URLParam(r, "workspaceID")
		workspaceID, err := uuid.Parse(workspaceIDStr)
		if err != nil {
			dto.ErrorResponse(w, http.StatusBadRequest, "invalid workspace ID")
			return
		}

		isMember, err := m.workspaceSvc.IsMember(r.Context(), workspaceID, claims.UserID)
		if err != nil {
			dto.ErrorResponse(w, http.StatusInternalServerError, "failed to check membership")
			return
		}
		if !isMember {
			dto.ErrorResponse(w, http.StatusForbidden, "not a member of this workspace")
			return
		}

		role, _ := m.workspaceSvc.GetMemberRole(r.Context(), workspaceID, claims.UserID)

		wsCtx := &WorkspaceContext{
			WorkspaceID: workspaceID,
			Role:        role,
		}
		ctx := context.WithValue(r.Context(), WorkspaceContextKey, wsCtx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *TenantMiddleware) RequireRole(requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wsCtx := GetWorkspaceFromContext(r.Context())
			if wsCtx == nil {
				dto.ErrorResponse(w, http.StatusForbidden, "workspace context required")
				return
			}

			if !hasRequiredRole(wsCtx.Role, requiredRole) {
				dto.ErrorResponse(w, http.StatusForbidden, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func GetWorkspaceFromContext(ctx context.Context) *WorkspaceContext {
	wsCtx, ok := ctx.Value(WorkspaceContextKey).(*WorkspaceContext)
	if !ok {
		return nil
	}
	return wsCtx
}

func hasRequiredRole(userRole, requiredRole string) bool {
	roleHierarchy := map[string]int{
		models.RoleViewer: 1,
		models.RoleMember: 2,
		models.RoleAdmin:  3,
		models.RoleOwner:  4,
	}

	return roleHierarchy[userRole] >= roleHierarchy[requiredRole]
}
