package marketplace

import (
	"net/http"
	"strconv"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

// BrowseHandler handles browse marketplace request
type BrowseHandler struct{}

// NewBrowseHandler creates a new handler
func NewBrowseHandler() *BrowseHandler {
	return &BrowseHandler{}
}

// Handle handles the browse marketplace request
func (h *BrowseHandler) Handle(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	category := r.URL.Query().Get("category")

	if limit <= 0 {
		limit = 20
	}

	templates := GetMarketplaceTemplates()
	if category != "" {
		filtered := []MarketplaceTemplate{}
		for _, t := range templates {
			if t.Category == category {
				filtered = append(filtered, t)
			}
		}
		templates = filtered
	}

	common.Success(w, map[string]interface{}{
		"templates": templates,
		"total":     len(templates),
		"limit":     limit,
		"offset":    offset,
	})
}
