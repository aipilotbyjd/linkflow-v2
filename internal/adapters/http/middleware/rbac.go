package middleware

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

// RequirePermission checks if the workspace member has the required permission
func RequirePermission(permID string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wsCtx := GetWorkspaceFromContext(r.Context())

			// If no workspace context, this middleware shouldn't be used or it's a server error
			if wsCtx == nil {
				common.Unauthorized(w, "Workspace context missing")
				return
			}

			// If no member (e.g. system admin or pure API key without member context?), usually member is set
			if wsCtx.Member == nil {
				common.Forbidden(w, "Member context missing")
				return
			}

			if !wsCtx.Member.HasPermission(permID) {
				common.Forbidden(w, "You do not have permission to perform this action")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
