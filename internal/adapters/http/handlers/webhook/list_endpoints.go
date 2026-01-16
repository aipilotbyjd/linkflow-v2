package webhook

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

type ListEndpointsHandler struct{}

func NewListEndpointsHandler() *ListEndpointsHandler {
	return &ListEndpointsHandler{}
}

func (h *ListEndpointsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "id")
	if workspaceID == "" {
		common.BadRequest(w, "Workspace ID is required")
		return
	}

	// TODO: Implement list endpoints
	// 1. Get all webhook endpoints for workspace
	// 2. Apply pagination

	common.Success(w, map[string]interface{}{
		"endpoints": []interface{}{},
		"total":     0,
		"page":      1,
		"page_size": 20,
	})
}
