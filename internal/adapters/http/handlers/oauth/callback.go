package oauth

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
)

// CallbackRequest represents OAuth callback request
type CallbackRequest struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

// CallbackResponse represents OAuth callback response
type CallbackResponse struct {
	CredentialID string `json:"credentialId"`
	Provider     string `json:"provider"`
}

// CallbackHandler handles OAuth callback request
type CallbackHandler struct {
	providers map[string]OAuthProvider
}

// NewCallbackHandler creates a new handler
func NewCallbackHandler(providers map[string]OAuthProvider) *CallbackHandler {
	return &CallbackHandler{providers: providers}
}

// Handle handles the OAuth callback request
func (h *CallbackHandler) Handle(w http.ResponseWriter, r *http.Request) {
	providerName := chi.URLParam(r, "provider")
	workspaceID := middleware.GetWorkspaceID(r.Context())

	provider, ok := h.providers[providerName]
	if !ok {
		common.BadRequest(w, "OAuth provider not supported")
		return
	}

	var req CallbackRequest

	if r.Method == http.MethodGet {
		req.Code = r.URL.Query().Get("code")
		req.State = r.URL.Query().Get("state")
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.BadRequest(w, "invalid request body")
			return
		}
	}

	if req.Code == "" {
		errorMsg := r.URL.Query().Get("error")
		if errorMsg != "" {
			common.BadRequest(w, errorMsg)
			return
		}
		common.BadRequest(w, "authorization code is required")
		return
	}

	token, err := provider.ExchangeCode(req.Code)
	if err != nil {
		common.BadRequest(w, "token exchange failed: "+err.Error())
		return
	}

	credentialID := uuid.New().String()
	_ = workspaceID
	_ = token

	common.Success(w, CallbackResponse{
		CredentialID: credentialID,
		Provider:     providerName,
	})
}
