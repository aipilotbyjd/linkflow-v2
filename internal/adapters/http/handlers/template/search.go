package template

import (
	"net/http"
	"strconv"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	templateQry "github.com/linkflow-ai/linkflow/internal/core/application/query/template"
)

type SearchHandler struct {
	handler *templateQry.SearchTemplatesHandler
}

func NewSearchHandler(handler *templateQry.SearchTemplatesHandler) *SearchHandler {
	return &SearchHandler{handler: handler}
}

func (h *SearchHandler) Handle(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	templates, total, err := h.handler.Handle(r.Context(), templateQry.SearchTemplatesQuery{
		Query:  query,
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
