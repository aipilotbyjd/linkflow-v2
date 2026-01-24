package rbac

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/rbac"
)

type ListPermissionsHandler struct {
	repo rbac.Repository
}

func NewListPermissionsHandler(repo rbac.Repository) *ListPermissionsHandler {
	return &ListPermissionsHandler{repo: repo}
}

func (h *ListPermissionsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	perms, err := h.repo.ListPermissions(r.Context())
	if err != nil {
		common.HandleError(w, err)
		return
	}

	res := make([]PermissionResponse, len(perms))
	for i, p := range perms {
		res[i] = PermissionResponse{
			ID:          p.ID,
			Scope:       p.Scope,
			Name:        p.Name,
			Description: p.Description,
		}
	}

	common.Success(w, map[string]interface{}{
		"permissions": res,
	})
}
