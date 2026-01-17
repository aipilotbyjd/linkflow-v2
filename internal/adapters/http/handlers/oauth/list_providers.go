package oauth

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

// ListProvidersHandler handles list OAuth providers request
type ListProvidersHandler struct {
	providers map[string]OAuthProvider
}

// NewListProvidersHandler creates a new handler
func NewListProvidersHandler(providers map[string]OAuthProvider) *ListProvidersHandler {
	return &ListProvidersHandler{providers: providers}
}

// Handle handles the list providers request
func (h *ListProvidersHandler) Handle(w http.ResponseWriter, r *http.Request) {
	providers := []ProviderInfo{
		{ID: "google", Name: "google", DisplayName: "Google"},
		{ID: "github", Name: "github", DisplayName: "GitHub"},
		{ID: "slack", Name: "slack", DisplayName: "Slack"},
		{ID: "microsoft", Name: "microsoft", DisplayName: "Microsoft"},
		{ID: "salesforce", Name: "salesforce", DisplayName: "Salesforce"},
		{ID: "hubspot", Name: "hubspot", DisplayName: "HubSpot"},
	}

	common.Success(w, map[string]interface{}{
		"providers": providers,
	})
}
