package template

import (
	"net/http"
	"strconv"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	templateQry "github.com/linkflow-ai/linkflow/internal/core/application/query/template"
)

type ListHandler struct {
	handler *templateQry.ListTemplatesHandler
}

func NewListHandler(handler *templateQry.ListTemplatesHandler) *ListHandler {
	return &ListHandler{handler: handler}
}

func (h *ListHandler) Handle(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	templates, total, err := h.handler.Handle(r.Context(), templateQry.ListTemplatesQuery{
		Limit:  limit,
		Offset: offset,
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
