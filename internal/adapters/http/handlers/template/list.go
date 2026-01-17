package template

import (
	"net/http"
	"strconv"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

// ListHandler handles list templates request
type ListHandler struct {
	repo TemplateRepository
}

// NewListHandler creates a new handler
func NewListHandler(repo TemplateRepository) *ListHandler {
	return &ListHandler{repo: repo}
}

// Handle handles the list templates request
func (h *ListHandler) Handle(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	templates := GetStaticTemplates()
	total := int64(len(templates))

	if offset < len(templates) {
		end := offset + limit
		if end > len(templates) {
			end = len(templates)
		}
		templates = templates[offset:end]
	} else {
		templates = []Template{}
	}

	common.Success(w, map[string]interface{}{
		"templates": templates,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}
