package middleware

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workspace"
)

// WorkspaceContext holds workspace context information
type WorkspaceContext struct {
	WorkspaceID uuid.UUID
	Workspace   *workspace.Workspace
	Member      *workspace.Member
	Role        workspace.Role
}

// Tenant creates a tenant (workspace) middleware
func Tenant(memberRepo workspace.MemberRepository, workspaceRepo workspace.Repository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get workspace ID from URL
			workspaceIDStr := chi.URLParam(r, "workspaceId")
			if workspaceIDStr == "" {
				common.BadRequest(w, "workspace ID is required")
				return
			}

			workspaceID, err := uuid.Parse(workspaceIDStr)
			if err != nil {
				common.BadRequest(w, "invalid workspace ID")
				return
			}

			// Get user from context
			user := GetUserFromContext(r.Context())
			if user == nil {
				common.Unauthorized(w, "authentication required")
				return
			}

			// Get workspace
			ws, err := workspaceRepo.FindByID(r.Context(), workspaceID)
			if err != nil {
				common.NotFound(w, "workspace")
				return
			}

			// Check membership
			member, err := memberRepo.FindByWorkspaceAndUser(r.Context(), workspaceID, user.UserID)
			if err != nil {
				common.Forbidden(w, "not a member of this workspace")
				return
			}

			// Set workspace context
			wsCtx := &WorkspaceContext{
				WorkspaceID: workspaceID,
				Workspace:   ws,
				Member:      member,
				Role:        member.Role,
			}
			ctx := context.WithValue(r.Context(), WorkspaceContextKey, wsCtx)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetWorkspaceFromContext retrieves workspace context
func GetWorkspaceFromContext(ctx context.Context) *WorkspaceContext {
	if wsCtx, ok := ctx.Value(WorkspaceContextKey).(*WorkspaceContext); ok {
		return wsCtx
	}
	return nil
}

// GetWorkspaceIDFromContext retrieves workspace ID from context
func GetWorkspaceIDFromContext(ctx context.Context) uuid.UUID {
	wsCtx := GetWorkspaceFromContext(ctx)
	if wsCtx != nil {
		return wsCtx.WorkspaceID
	}
	return uuid.Nil
}

// GetWorkspaceID is an alias for GetWorkspaceIDFromContext
func GetWorkspaceID(ctx context.Context) uuid.UUID {
	return GetWorkspaceIDFromContext(ctx)
}

// RequireWorkspace ensures workspace context is present
func RequireWorkspace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if GetWorkspaceFromContext(r.Context()) == nil {
			common.BadRequest(w, "workspace context required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireRole creates middleware that requires a minimum role
func RequireRole(minRole workspace.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wsCtx := GetWorkspaceFromContext(r.Context())
			if wsCtx == nil {
				common.BadRequest(w, "workspace context required")
				return
			}

			if wsCtx.Role.Level() < minRole.Level() {
				common.Forbidden(w, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin ensures user has admin role
func RequireAdmin(next http.Handler) http.Handler {
	return RequireRole(workspace.RoleAdmin)(next)
}

// RequireOwner ensures user is workspace owner
func RequireOwner(next http.Handler) http.Handler {
	return RequireRole(workspace.RoleOwner)(next)
}
