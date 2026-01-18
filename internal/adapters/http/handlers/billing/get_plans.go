package billing

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	billingQuery "github.com/linkflow-ai/linkflow/internal/core/application/query/billing"
)

type GetPlansHandler struct {
	handler *billingQuery.GetPlansHandler
}

func NewGetPlansHandler(handler *billingQuery.GetPlansHandler) *GetPlansHandler {
	return &GetPlansHandler{handler: handler}
}

func (h *GetPlansHandler) Handle(w http.ResponseWriter, r *http.Request) {
	plans := h.handler.Handle()
	common.Success(w, map[string]interface{}{
		"plans": ToPlanResponses(plans),
	})
}
