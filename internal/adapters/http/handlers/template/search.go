package template

import (
	"net/http"
	"strconv"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

// SearchHandler handles search templates request
type SearchHandler struct {
	repo TemplateRepository
}

// NewSearchHandler creates a new handler
func NewSearchHandler(repo TemplateRepository) *SearchHandler {
	return &SearchHandler{repo: repo}
}

// Handle handles the search templates request
func (h *SearchHandler) Handle(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 20
	}

	templates := GetStaticTemplates()

	common.Success(w, map[string]interface{}{
		"templates": templates,
		"query":     query,
		"total":     len(templates),
		"limit":     limit,
		"offset":    offset,
	})
}
