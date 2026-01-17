package template

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

// ByCategoryHandler handles get templates by category request
type ByCategoryHandler struct {
	repo TemplateRepository
}

// NewByCategoryHandler creates a new handler
func NewByCategoryHandler(repo TemplateRepository) *ByCategoryHandler {
	return &ByCategoryHandler{repo: repo}
}

// Handle handles the get templates by category request
func (h *ByCategoryHandler) Handle(w http.ResponseWriter, r *http.Request) {
	category := chi.URLParam(r, "category")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 20
	}

	templates := GetStaticTemplates()
	filtered := []Template{}
	for _, t := range templates {
		if t.Category == category {
			filtered = append(filtered, t)
		}
	}

	total := int64(len(filtered))
	if offset < len(filtered) {
		end := offset + limit
		if end > len(filtered) {
			end = len(filtered)
		}
		filtered = filtered[offset:end]
	} else {
		filtered = []Template{}
	}

	common.Success(w, map[string]interface{}{
		"templates": filtered,
		"category":  category,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}
