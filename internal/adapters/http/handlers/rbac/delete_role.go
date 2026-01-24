package rbac

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/rbac"
)

type DeleteRoleHandler struct {
	repo rbac.Repository
}

func NewDeleteRoleHandler(repo rbac.Repository) *DeleteRoleHandler {
	return &DeleteRoleHandler{repo: repo}
}

func (h *DeleteRoleHandler) Handle(w http.ResponseWriter, r *http.Request) {
	roleIDStr := chi.URLParam(r, "roleId")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		common.BadRequest(w, "invalid role ID")
		return
	}

	workspaceID := middleware.GetWorkspaceID(r.Context())

	// Fetch Role
	role, err := h.repo.GetRole(r.Context(), roleID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	// Check scope
	if role.IsSystem() {
		common.Forbidden(w, "Cannot delete system roles")
		return
	}
	if role.WorkspaceID == nil || *role.WorkspaceID != workspaceID {
		common.NotFound(w, "Role not found in this workspace")
		return
	}

	if role.IsProtected {
		common.Forbidden(w, "Cannot delete protected roles")
		return
	}

	// TODO: Check if role is assigned to any members before deleting?
	// The DB constraint might handle this or we might want a soft check.
	// For now, let DB constraint fail if assigned.

	if err := h.repo.DeleteRole(r.Context(), roleID); err != nil {
		common.HandleError(w, err)
		return
	}

	common.NoContent(w)
}
