package rbac

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/rbac"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/validation"
)

type CreateRoleHandler struct {
	repo rbac.Repository
}

func NewCreateRoleHandler(repo rbac.Repository) *CreateRoleHandler {
	return &CreateRoleHandler{repo: repo}
}

func (h *CreateRoleHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workspaceID := middleware.GetWorkspaceID(r.Context())

	var req CreateRoleRequest
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

	// Check if role name already exists in workspace
	existing, err := h.repo.GetRoleByName(r.Context(), &workspaceID, req.Name)
	if err == nil && existing != nil {
		common.BadRequest(w, "Role with this name already exists")
		return
	}

	role := rbac.NewRole(&workspaceID, req.Name, req.Description)

	if err := h.repo.CreateRole(r.Context(), role); err != nil {
		common.HandleError(w, err)
		return
	}

	if len(req.Permissions) > 0 {
		if err := h.repo.AssignPermissions(r.Context(), role.ID, req.Permissions); err != nil {
			common.HandleError(w, err)
			return
		}
		// Refresh role to get permissions
		role, _ = h.repo.GetRole(r.Context(), role.ID)
	}

	common.Created(w, ToRoleResponse(role))
}
