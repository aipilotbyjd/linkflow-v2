package oauth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

// AuthorizeHandler handles OAuth authorization redirect
type AuthorizeHandler struct {
	providers map[string]OAuthProvider
}

// NewAuthorizeHandler creates a new handler
func NewAuthorizeHandler(providers map[string]OAuthProvider) *AuthorizeHandler {
	return &AuthorizeHandler{providers: providers}
}

// Handle handles the OAuth authorization request
func (h *AuthorizeHandler) Handle(w http.ResponseWriter, r *http.Request) {
	providerID := chi.URLParam(r, "provider")

	provider, ok := h.providers[providerID]
	if !ok {
		common.NotFound(w, "OAuth provider")
		return
	}

	// Generate state for CSRF protection
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		common.HandleError(w, err)
		return
	}
	state := hex.EncodeToString(stateBytes)

	// TODO: Store state in session/cache for validation in callback

	authURL := provider.GetAuthURL(state)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}
