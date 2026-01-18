package oauth

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

// CallbackHandler handles OAuth callback
type CallbackHandler struct {
	providers map[string]OAuthProvider
}

// NewCallbackHandler creates a new handler
func NewCallbackHandler(providers map[string]OAuthProvider) *CallbackHandler {
	return &CallbackHandler{providers: providers}
}

// Handle handles the OAuth callback request
func (h *CallbackHandler) Handle(w http.ResponseWriter, r *http.Request) {
	providerID := chi.URLParam(r, "provider")

	provider, ok := h.providers[providerID]
	if !ok {
		common.NotFound(w, "OAuth provider")
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		errorMsg := r.URL.Query().Get("error")
		if errorMsg != "" {
			common.BadRequest(w, "OAuth error: "+errorMsg)
			return
		}
		common.BadRequest(w, "Authorization code not provided")
		return
	}

	// TODO: Validate state parameter

	// Exchange code for token
	token, err := provider.ExchangeCode(code)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	// Get user info
	userInfo, err := provider.GetUserInfo(token)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	// TODO: Create/update user and generate JWT tokens
	// For now, return the user info
	common.Success(w, map[string]interface{}{
		"provider": providerID,
		"user":     userInfo,
		"message":  "OAuth authentication successful",
	})
}
