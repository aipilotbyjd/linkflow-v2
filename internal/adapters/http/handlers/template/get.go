package template

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

// GetHandler handles get template request
type GetHandler struct {
	repo TemplateRepository
}

// NewGetHandler creates a new handler
func NewGetHandler(repo TemplateRepository) *GetHandler {
	return &GetHandler{repo: repo}
}

// Handle handles the get template request
func (h *GetHandler) Handle(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateId")

	templates := GetStaticTemplates()
	for _, t := range templates {
		if t.ID == templateID {
			common.Success(w, t)
			return
		}
	}

	common.NotFound(w, "template")
}
