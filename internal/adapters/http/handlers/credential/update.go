package credential

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	credentialCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/credential"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/validation"
)

type UpdateHandler struct {
	handler *credentialCmd.UpdateCredentialHandler
}

func NewUpdateHandler(handler *credentialCmd.UpdateCredentialHandler) *UpdateHandler {
	return &UpdateHandler{handler: handler}
}

func (h *UpdateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	credentialIDStr := chi.URLParam(r, "credentialId")
	credentialID, err := uuid.Parse(credentialIDStr)
	if err != nil {
		common.BadRequest(w, "invalid credential ID")
		return
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}

	if errors := validation.Validate(req); len(errors) > 0 {
		details := make([]common.ValidationDetail, len(errors))
		for i, e := range errors {
			details[i] = common.ValidationDetail{Field: e.Field, Message: e.Message}
		}
		common.ValidationErrors(w, details)
		return
	}

	cred, err := h.handler.Handle(r.Context(), credentialCmd.UpdateCredentialCommand{
		CredentialID: credentialID,
		Name:         req.Name,
		Description:  req.Description,
		Data:         req.Data,
		SharingScope: req.Scope,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, ToCredentialResponse(cred))
}
