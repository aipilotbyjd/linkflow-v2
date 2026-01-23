package oauth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"time"

	"github.com/go-chi/chi/v5"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/cache"
)

// AuthorizeHandler handles OAuth authorization redirect
type AuthorizeHandler struct {
	providers map[string]OAuthProvider
	cache     cache.Cache
}

// NewAuthorizeHandler creates a new handler
func NewAuthorizeHandler(providers map[string]OAuthProvider, cache cache.Cache) *AuthorizeHandler {
	return &AuthorizeHandler{
		providers: providers,
		cache:     cache,
	}
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

	// Store state in cache for validation in callback (10 minutes TTL)
	if err := h.cache.Set(r.Context(), "oauth_state:"+state, []byte(providerID), 10*time.Minute); err != nil {
		common.HandleError(w, err)
		return
	}

	authURL := provider.GetAuthURL(state)

	// Return JSON with authorization URL instead of redirect
	common.Success(w, map[string]interface{}{
		"provider": providerID,
		"auth_url": authURL,
		"state":    state,
	})
}
