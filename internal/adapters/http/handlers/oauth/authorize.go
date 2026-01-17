package oauth

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

// AuthorizeResponse represents authorize response
type AuthorizeResponse struct {
	AuthURL string `json:"authUrl"`
	State   string `json:"state"`
}

// AuthorizeHandler handles OAuth authorize request
type AuthorizeHandler struct {
	providers map[string]OAuthProvider
}

// NewAuthorizeHandler creates a new handler
func NewAuthorizeHandler(providers map[string]OAuthProvider) *AuthorizeHandler {
	return &AuthorizeHandler{providers: providers}
}

// Handle handles the authorize request
func (h *AuthorizeHandler) Handle(w http.ResponseWriter, r *http.Request) {
	providerName := chi.URLParam(r, "provider")

	provider, ok := h.providers[providerName]
	if !ok {
		common.BadRequest(w, "OAuth provider not supported")
		return
	}

	state := uuid.New().String()
	authURL := provider.GetAuthURL(state)

	common.Success(w, AuthorizeResponse{
		AuthURL: authURL,
		State:   state,
	})
}
