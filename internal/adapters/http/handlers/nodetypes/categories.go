package nodetypes

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/worker/nodes"
)

type ListCategoriesHandler struct {
	registry *nodes.Registry
}

func NewListCategoriesHandler(registry *nodes.Registry) *ListCategoriesHandler {
	return &ListCategoriesHandler{registry: registry}
}

type CategoryResponse struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func (h *ListCategoriesHandler) Handle(w http.ResponseWriter, r *http.Request) {
	byCategory := h.registry.ListByCategory()

	categories := make([]CategoryResponse, 0, len(byCategory))
	for name, nodeTypes := range byCategory {
		categories = append(categories, CategoryResponse{
			Name:  name,
			Count: len(nodeTypes),
		})
	}

	common.Success(w, categories)
}
