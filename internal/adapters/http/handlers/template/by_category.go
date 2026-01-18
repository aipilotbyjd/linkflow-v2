package template

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	templateQry "github.com/linkflow-ai/linkflow/internal/core/application/query/template"
)

type ByCategoryHandler struct {
	handler *templateQry.GetByCategoryHandler
}

func NewByCategoryHandler(handler *templateQry.GetByCategoryHandler) *ByCategoryHandler {
	return &ByCategoryHandler{handler: handler}
}

func (h *ByCategoryHandler) Handle(w http.ResponseWriter, r *http.Request) {
	category := chi.URLParam(r, "category")
	if category == "" {
		common.BadRequest(w, "category is required")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	templates, total, err := h.handler.Handle(r.Context(), templateQry.GetByCategoryQuery{
		Category: category,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, map[string]interface{}{
		"templates": ToTemplateResponses(templates),
		"total":     total,
	})
}
