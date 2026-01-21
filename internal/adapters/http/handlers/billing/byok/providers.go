package byok

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
)

// ProvidersHandler lists supported AI providers
type ProvidersHandler struct{}

func NewProvidersHandler() *ProvidersHandler {
	return &ProvidersHandler{}
}

func (h *ProvidersHandler) Handle(w http.ResponseWriter, r *http.Request) {
	providers := billing.SupportedProviders

	responses := make([]ProviderResponse, len(providers))
	for i, p := range providers {
		responses[i] = ToProviderResponse(p)
	}

	common.Success(w, map[string]interface{}{
		"providers": responses,
	})
}
