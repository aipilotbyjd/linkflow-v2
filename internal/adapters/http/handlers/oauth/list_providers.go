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
	providerList := make([]ProviderInfo, 0, len(h.providers))

	for id, provider := range h.providers {
		displayName := ProviderDisplayNames[id]
		if displayName == "" {
			displayName = provider.Name()
		}

		providerList = append(providerList, ProviderInfo{
			ID:          id,
			Name:        provider.Name(),
			DisplayName: displayName,
			Icon:        ProviderIcons[id],
		})
	}

	common.Success(w, map[string]interface{}{
		"providers": providerList,
	})
}
