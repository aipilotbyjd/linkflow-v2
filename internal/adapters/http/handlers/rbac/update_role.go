package rbac

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/rbac"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/validation"
)

type UpdateRoleHandler struct {
	repo rbac.Repository
}

func NewUpdateRoleHandler(repo rbac.Repository) *UpdateRoleHandler {
	return &UpdateRoleHandler{repo: repo}
}

func (h *UpdateRoleHandler) Handle(w http.ResponseWriter, r *http.Request) {
	roleIDStr := chi.URLParam(r, "roleId")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		common.BadRequest(w, "invalid role ID")
		return
	}

	workspaceID := middleware.GetWorkspaceID(r.Context())

	var req UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}

	if errors := validation.Validate(req); len(errors) > 0 {
		details := make([]common.ValidationDetail, len(errors))
		for i, e := range errors {
			details[i] = common.ValidationDetail{Field: e.Field, Message: e.Message}
		}
		common.ValidationErrors(w, details)
		return
	}

	// Fetch Role
	role, err := h.repo.GetRole(r.Context(), roleID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	// Check scope (cannot update system roles or roles from other workspaces)
	if role.IsSystem() {
		common.Forbidden(w, "Cannot modify system roles")
		return
	}
	if role.WorkspaceID == nil || *role.WorkspaceID != workspaceID {
		common.NotFound(w, "Role not found in this workspace")
		return
	}

	// Update fields
	if req.Name != "" {
		// Check name uniqueness if changed
		if req.Name != role.Name {
			existing, err := h.repo.GetRoleByName(r.Context(), &workspaceID, req.Name)
			if err == nil && existing != nil {
				common.BadRequest(w, "Role with this name already exists")
				return
			}
		}
		role.Name = req.Name
	}
	if req.Description != "" {
		role.Description = req.Description
	}

	if err := h.repo.UpdateRole(r.Context(), role); err != nil {
		common.HandleError(w, err)
		return
	}

	// Update permissions if provided
	if req.Permissions != nil {
		if err := h.repo.AssignPermissions(r.Context(), role.ID, req.Permissions); err != nil {
			common.HandleError(w, err)
			return
		}
		// Refresh permissions
		role, _ = h.repo.GetRole(r.Context(), role.ID)
	}

	common.Success(w, ToRoleResponse(role))
}
