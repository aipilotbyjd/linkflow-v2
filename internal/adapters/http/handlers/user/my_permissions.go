package user

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
)

type MyPermissionsHandler struct {
	// No repo needed if we get everything from context
}

func NewMyPermissionsHandler() *MyPermissionsHandler {
	return &MyPermissionsHandler{}
}

func (h *MyPermissionsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil || wsCtx.Member == nil {
		common.Unauthorized(w, "Workspace context required")
		return
	}

	// Extract permissions from Member's RBAC Role
	// If RBACRole is loaded
	var permissions []string
	if wsCtx.Member.RBACRole != nil {
		permissions = make([]string, len(wsCtx.Member.RBACRole.Permissions))
		for i, p := range wsCtx.Member.RBACRole.Permissions {
			permissions[i] = p.ID
		}
	} else {
		// Fallback or empty if no role (shouldn't happen with correct middleware)
		permissions = []string{}
	}

	common.Success(w, map[string]interface{}{
		"role":        wsCtx.Member.Role, // Legacy role name
		"rbac_role":   wsCtx.Member.RBACRole,
		"permissions": permissions,
	})
}
