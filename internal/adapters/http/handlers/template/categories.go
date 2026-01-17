package template

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

// CategoriesHandler handles get categories request
type CategoriesHandler struct {
	repo TemplateRepository
}

// NewCategoriesHandler creates a new handler
func NewCategoriesHandler(repo TemplateRepository) *CategoriesHandler {
	return &CategoriesHandler{repo: repo}
}

// Handle handles the get categories request
func (h *CategoriesHandler) Handle(w http.ResponseWriter, r *http.Request) {
	categories := []TemplateCategory{
		{ID: "automation", Name: "Automation", Description: "Automate repetitive tasks", Count: 25, Icon: "robot"},
		{ID: "integration", Name: "Integration", Description: "Connect different services", Count: 42, Icon: "link"},
		{ID: "data", Name: "Data Processing", Description: "Transform and analyze data", Count: 18, Icon: "database"},
		{ID: "communication", Name: "Communication", Description: "Email, Slack, and messaging", Count: 15, Icon: "message"},
		{ID: "crm", Name: "CRM & Sales", Description: "Sales and customer management", Count: 12, Icon: "users"},
		{ID: "devops", Name: "DevOps", Description: "CI/CD and infrastructure", Count: 8, Icon: "code"},
	}

	common.Success(w, map[string]interface{}{
		"categories": categories,
	})
}
