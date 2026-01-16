package credential

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	credentialCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/credential"
)

type DeleteHandler struct {
	handler *credentialCmd.DeleteCredentialHandler
}

func NewDeleteHandler(handler *credentialCmd.DeleteCredentialHandler) *DeleteHandler {
	return &DeleteHandler{handler: handler}
}

func (h *DeleteHandler) Handle(w http.ResponseWriter, r *http.Request) {
	credentialIDStr := chi.URLParam(r, "credentialId")
	credentialID, err := uuid.Parse(credentialIDStr)
	if err != nil {
		common.BadRequest(w, "invalid credential ID")
		return
	}

	if err := h.handler.Handle(r.Context(), credentialCmd.DeleteCredentialCommand{
		CredentialID: credentialID,
	}); err != nil {
		common.HandleError(w, err)
		return
	}

	common.NoContent(w)
}
