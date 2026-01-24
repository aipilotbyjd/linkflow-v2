package rbac

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/rbac"
)

type ListRolesHandler struct {
	repo rbac.Repository
}

func NewListRolesHandler(repo rbac.Repository) *ListRolesHandler {
	return &ListRolesHandler{repo: repo}
}

func (h *ListRolesHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workspaceID := middleware.GetWorkspaceID(r.Context())

	roles, err := h.repo.ListRoles(r.Context(), workspaceID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, map[string]interface{}{
		"roles": ToRoleResponseList(roles),
	})
}
