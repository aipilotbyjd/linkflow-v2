package marketplace

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

// UpdateHandler handles update marketplace template request
type UpdateHandler struct{}

// NewUpdateHandler creates a new handler
func NewUpdateHandler() *UpdateHandler {
	return &UpdateHandler{}
}

// Handle handles the update request
func (h *UpdateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateId")

	var req PublishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "invalid request body")
		return
	}

	template := MarketplaceTemplate{
		ID:          templateID,
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Tags:        req.Tags,
		Price:       req.Price,
		UpdatedAt:   time.Now(),
	}

	common.Success(w, template)
}
