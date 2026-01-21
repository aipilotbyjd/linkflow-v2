package topup

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
)

// ListPacksHandler lists available credit packs
type ListPacksHandler struct{}

func NewListPacksHandler() *ListPacksHandler {
	return &ListPacksHandler{}
}

func (h *ListPacksHandler) Handle(w http.ResponseWriter, r *http.Request) {
	packs := billing.CreditPacks

	responses := make([]CreditPackResponse, len(packs))
	for i, pack := range packs {
		responses[i] = ToCreditPackResponse(pack)
	}

	common.Success(w, map[string]interface{}{
		"packs": responses,
	})
}
