package marketplace

import (
	"net/http"
	"strconv"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

// FeaturedHandler handles get featured marketplace templates request
type FeaturedHandler struct{}

// NewFeaturedHandler creates a new handler
func NewFeaturedHandler() *FeaturedHandler {
	return &FeaturedHandler{}
}

// Handle handles the get featured templates request
func (h *FeaturedHandler) Handle(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 6
	}

	templates := GetMarketplaceTemplates()
	featured := []MarketplaceTemplate{}
	for _, t := range templates {
		if t.Featured && len(featured) < limit {
			featured = append(featured, t)
		}
	}

	common.Success(w, map[string]interface{}{
		"templates": featured,
	})
}
