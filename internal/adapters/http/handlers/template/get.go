package template

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	templateQry "github.com/linkflow-ai/linkflow/internal/core/application/query/template"
)

type GetHandler struct {
	handler *templateQry.GetTemplateHandler
}

func NewGetHandler(handler *templateQry.GetTemplateHandler) *GetHandler {
	return &GetHandler{handler: handler}
}

func (h *GetHandler) Handle(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "templateId")
	id, err := uuid.Parse(idStr)
	if err != nil {
		common.BadRequest(w, "invalid template ID")
		return
	}

	tmpl, err := h.handler.Handle(r.Context(), templateQry.GetTemplateQuery{
		TemplateID: id,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	if tmpl == nil {
		common.NotFound(w, "template")
		return
	}

	common.Success(w, ToTemplateDetailResponse(tmpl))
}
