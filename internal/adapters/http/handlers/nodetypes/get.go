package nodetypes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/worker/nodes"
)

type GetNodeTypeHandler struct {
	registry *nodes.Registry
}

func NewGetNodeTypeHandler(registry *nodes.Registry) *GetNodeTypeHandler {
	return &GetNodeTypeHandler{registry: registry}
}

func (h *GetNodeTypeHandler) Handle(w http.ResponseWriter, r *http.Request) {
	nodeType := chi.URLParam(r, "nodeType")
	if nodeType == "" {
		common.BadRequest(w, "node type is required")
		return
	}

	metadata, err := h.registry.GetMetadata(nodeType)
	if err != nil {
		common.NotFound(w, "node type")
		return
	}

	common.Success(w, metadata)
}
