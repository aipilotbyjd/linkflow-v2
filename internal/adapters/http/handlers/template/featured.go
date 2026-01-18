package template

import (
	"net/http"
	"strconv"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	templateQry "github.com/linkflow-ai/linkflow/internal/core/application/query/template"
)

type FeaturedHandler struct {
	handler *templateQry.GetFeaturedHandler
}

func NewFeaturedHandler(handler *templateQry.GetFeaturedHandler) *FeaturedHandler {
	return &FeaturedHandler{handler: handler}
}

func (h *FeaturedHandler) Handle(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 10
	}

	templates, err := h.handler.Handle(r.Context(), templateQry.GetFeaturedQuery{
		Limit: limit,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, map[string]interface{}{
		"templates": ToTemplateResponses(templates),
	})
}
