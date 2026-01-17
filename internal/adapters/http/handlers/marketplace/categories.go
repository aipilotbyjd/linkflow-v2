package marketplace

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

// CategoriesHandler handles get marketplace categories request
type CategoriesHandler struct{}

// NewCategoriesHandler creates a new handler
func NewCategoriesHandler() *CategoriesHandler {
	return &CategoriesHandler{}
}

// Handle handles the get categories request
func (h *CategoriesHandler) Handle(w http.ResponseWriter, r *http.Request) {
	categories := []map[string]interface{}{
		{"id": "automation", "name": "Automation", "count": 45, "icon": "robot"},
		{"id": "integration", "name": "Integration", "count": 78, "icon": "link"},
		{"id": "data", "name": "Data Processing", "count": 32, "icon": "database"},
		{"id": "communication", "name": "Communication", "count": 56, "icon": "message"},
		{"id": "crm", "name": "CRM & Sales", "count": 29, "icon": "users"},
		{"id": "devops", "name": "DevOps", "count": 41, "icon": "code"},
		{"id": "ai", "name": "AI & ML", "count": 23, "icon": "brain"},
	}

	common.Success(w, map[string]interface{}{
		"categories": categories,
	})
}
