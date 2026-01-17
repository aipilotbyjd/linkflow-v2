package marketplace

import (
	"net/http"
	"strconv"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

// SearchHandler handles search marketplace request
type SearchHandler struct{}

// NewSearchHandler creates a new handler
func NewSearchHandler() *SearchHandler {
	return &SearchHandler{}
}

// Handle handles the search marketplace request
func (h *SearchHandler) Handle(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 20
	}

	templates := GetMarketplaceTemplates()

	common.Success(w, map[string]interface{}{
		"templates": templates,
		"query":     query,
		"total":     len(templates),
		"limit":     limit,
		"offset":    offset,
	})
}
