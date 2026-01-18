package credential

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	credentialCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/credential"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/validation"
)

// CreateHandler handles credential creation
type CreateHandler struct {
	handler *credentialCmd.CreateCredentialHandler
}

// NewCreateHandler creates a new handler
func NewCreateHandler(handler *credentialCmd.CreateCredentialHandler) *CreateHandler {
	return &CreateHandler{handler: handler}
}

// Handle handles the create credential request
func (h *CreateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
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

	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		common.BadRequest(w, "workspace context required")
		return
	}

	userClaims := middleware.GetUserFromContext(r.Context())
	if userClaims == nil {
		common.Unauthorized(w, "")
		return
	}

	cred, err := h.handler.Handle(r.Context(), credentialCmd.CreateCredentialCommand{
		WorkspaceID:  wsCtx.WorkspaceID,
		CreatedBy:    userClaims.UserID,
		Name:         req.Name,
		Description:  req.Description,
		Type:         req.Type,
		Provider:     req.Provider,
		Data:         req.Data,
		SharingScope: req.Scope,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Created(w, ToCredentialResponse(cred))
}
