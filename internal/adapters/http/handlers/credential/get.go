package credential

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	credentialQuery "github.com/linkflow-ai/linkflow/internal/core/application/query/credential"
)

type GetHandler struct {
	handler *credentialQuery.GetCredentialHandler
}

func NewGetHandler(handler *credentialQuery.GetCredentialHandler) *GetHandler {
	return &GetHandler{handler: handler}
}

func (h *GetHandler) Handle(w http.ResponseWriter, r *http.Request) {
	credentialIDStr := chi.URLParam(r, "credentialId")
	credentialID, err := uuid.Parse(credentialIDStr)
	if err != nil {
		common.BadRequest(w, "invalid credential ID")
		return
	}

	cred, err := h.handler.Handle(r.Context(), credentialQuery.GetCredentialQuery{
		CredentialID: credentialID,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, toCredentialResponse(cred))
}
