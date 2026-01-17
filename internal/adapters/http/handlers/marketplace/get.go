package marketplace

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

// GetHandler handles get marketplace template request
type GetHandler struct{}

// NewGetHandler creates a new handler
func NewGetHandler() *GetHandler {
	return &GetHandler{}
}

// Handle handles the get template request
func (h *GetHandler) Handle(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateId")

	templates := GetMarketplaceTemplates()
	for _, t := range templates {
		if t.ID == templateID {
			t.LongDescription = "This is a detailed description of the template with usage instructions and examples."
			t.Screenshots = []string{"/screenshots/1.png", "/screenshots/2.png"}
			common.Success(w, t)
			return
		}
	}

	common.NotFound(w, "template")
}
