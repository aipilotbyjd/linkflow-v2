package auth

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

type OAuthManager interface {
	GetAuthURL(provider, state, redirectURL string) (string, error)
	ExchangeCode(provider, code string) (*OAuthToken, error)
}

type OAuthToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type OAuthRedirectHandler struct {
	oauthManager OAuthManager
	frontendURL  string
}

func NewOAuthRedirectHandler(oauthManager OAuthManager, frontendURL string) *OAuthRedirectHandler {
	return &OAuthRedirectHandler{
		oauthManager: oauthManager,
		frontendURL:  frontendURL,
	}
}

func (h *OAuthRedirectHandler) Handle(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	
	validProviders := map[string]bool{
		"google":   true,
		"github":   true,
		"slack":    true,
		"microsoft": true,
	}
	
	if !validProviders[provider] {
		common.Error(w, http.StatusBadRequest, "INVALID_PROVIDER", "OAuth provider not supported")
		return
	}

	state := uuid.New().String()
	
	redirectURL := r.URL.Query().Get("redirect_uri")
	if redirectURL == "" {
		redirectURL = h.frontendURL + "/auth/callback"
	}

	var authURL string
	if h.oauthManager != nil {
		var err error
		authURL, err = h.oauthManager.GetAuthURL(provider, state, redirectURL)
		if err != nil {
			common.Error(w, http.StatusInternalServerError, "OAUTH_ERROR", "Failed to generate OAuth URL")
			return
		}
	} else {
		authURL = "https://accounts.google.com/o/oauth2/v2/auth?state=" + state
	}

	common.Success(w, map[string]string{
		"authUrl":  authURL,
		"provider": provider,
		"state":    state,
	})
}
