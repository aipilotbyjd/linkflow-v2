package template

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

type CategoriesHandler struct{}

func NewCategoriesHandler() *CategoriesHandler {
	return &CategoriesHandler{}
}

func (h *CategoriesHandler) Handle(w http.ResponseWriter, r *http.Request) {
	common.Success(w, map[string]interface{}{
		"categories": GetCategories(),
	})
}
