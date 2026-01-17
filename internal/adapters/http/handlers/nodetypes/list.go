package nodetypes

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/worker/nodes"
)

type ListNodeTypesHandler struct {
	registry *nodes.Registry
}

func NewListNodeTypesHandler(registry *nodes.Registry) *ListNodeTypesHandler {
	return &ListNodeTypesHandler{registry: registry}
}

func (h *ListNodeTypesHandler) Handle(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")

	if category != "" {
		// Filter by category
		byCategory := h.registry.ListByCategory()
		if nodeTypes, ok := byCategory[category]; ok {
			common.Success(w, nodeTypes)
			return
		}
		common.Success(w, []interface{}{})
		return
	}

	// Return all node types
	metadata := h.registry.ListMetadata()
	common.Success(w, metadata)
}
