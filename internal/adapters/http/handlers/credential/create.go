package credential

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	credentialCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/credential"
	"github.com/linkflow-ai/linkflow/internal/core/domain/credential"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// CreateRequest represents credential creation request
type CreateRequest struct {
	Name        string               `json:"name" validate:"required"`
	Description *string              `json:"description,omitempty"`
	Type        credential.Type      `json:"type" validate:"required"`
	Provider    string               `json:"provider"`
	Data        types.JSON           `json:"data" validate:"required"`
	Scope       credential.SharingScope `json:"scope"`
}

// CredentialResponse represents credential in responses
type CredentialResponse struct {
	ID          string  `json:"id"`
	WorkspaceID string  `json:"workspace_id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Type        string  `json:"type"`
	Provider    string  `json:"provider"`
	Scope       string  `json:"scope"`
	CreatedBy   string  `json:"created_by"`
	LastUsedAt  *string `json:"last_used_at,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

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
		common.BadRequest(w, "invalid request body")
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

	common.Created(w, toCredentialResponse(cred))
}

func toCredentialResponse(c *credential.Credential) CredentialResponse {
	provider := ""
	if c.Provider != nil {
		provider = *c.Provider
	}
	resp := CredentialResponse{
		ID:          c.ID.String(),
		WorkspaceID: c.WorkspaceID.String(),
		Name:        c.Name,
		Description: c.Description,
		Type:        string(c.Type),
		Provider:    provider,
		Scope:       string(c.SharingScope),
		CreatedBy:   c.CreatedBy.String(),
		CreatedAt:   c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   c.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if c.LastUsedAt != nil {
		s := c.LastUsedAt.Format("2006-01-02T15:04:05Z")
		resp.LastUsedAt = &s
	}

	return resp
}
