package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/rs/zerolog/log"
)

type OAuthHandler struct {
	oauthSvc *services.OAuthService
}

func NewOAuthHandler(oauthSvc *services.OAuthService) *OAuthHandler {
	return &OAuthHandler{oauthSvc: oauthSvc}
}

// GetProviders returns list of supported OAuth providers
func (h *OAuthHandler) GetProviders(w http.ResponseWriter, r *http.Request) {
	providers := h.oauthSvc.GetSupportedProviders()
	dto.NewResponse(providers).
		WithLinks(&dto.Links{Self: "/api/v1/oauth/providers"}).
		Send(w)
}

// AuthorizeRequest represents OAuth authorization request
type AuthorizeRequest struct {
	Provider       string   `json:"provider"`
	CredentialName string   `json:"credential_name,omitempty"`
	Scopes         []string `json:"scopes,omitempty"`
	RedirectURL    string   `json:"redirect_url,omitempty"`
}

// Authorize initiates OAuth flow - returns authorization URL
func (h *OAuthHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")

	// Also support POST with body
	var req AuthorizeRequest
	if r.Method == http.MethodPost {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			dto.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.Provider != "" {
			provider = req.Provider
		}
	} else {
		// GET request - read from query params
		req.CredentialName = r.URL.Query().Get("credential_name")
		req.RedirectURL = r.URL.Query().Get("redirect_url")
		if scopesParam := r.URL.Query().Get("scopes"); scopesParam != "" {
			req.Scopes = []string{scopesParam} // Simple single scope for now
		}
	}

	if provider == "" {
		dto.ErrorResponse(w, http.StatusBadRequest, "provider is required")
		return
	}

	userCtx, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if userCtx == nil || wsCtx == nil {
		return
	}

	result, err := h.oauthSvc.GetAuthorizationURL(r.Context(), services.AuthURLInput{
		Provider:       provider,
		UserID:         userCtx.UserID,
		WorkspaceID:    wsCtx.WorkspaceID,
		CredentialName: req.CredentialName,
		Scopes:         req.Scopes,
		RedirectURL:    req.RedirectURL,
	})
	if err != nil {
		log.Warn().Err(err).Str("provider", provider).Msg("Failed to get OAuth authorization URL")
		dto.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, result)
}

// Callback handles OAuth callback from provider
// This is called by the OAuth provider, not the frontend
func (h *OAuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errParam := r.URL.Query().Get("error")
	errDesc := r.URL.Query().Get("error_description")

	frontendURL := h.oauthSvc.GetFrontendURL()

	// Handle OAuth error from provider
	if errParam != "" {
		log.Warn().
			Str("provider", provider).
			Str("error", errParam).
			Str("description", errDesc).
			Msg("OAuth provider returned error")

		redirectURL := fmt.Sprintf("%s/credentials?oauth=error&error=%s&error_description=%s",
			frontendURL,
			url.QueryEscape(errParam),
			url.QueryEscape(errDesc))
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	if code == "" || state == "" {
		redirectURL := fmt.Sprintf("%s/credentials?oauth=error&error=missing_params&error_description=%s",
			frontendURL,
			url.QueryEscape("Missing code or state parameter"))
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	result, err := h.oauthSvc.HandleCallback(r.Context(), services.CallbackInput{
		Provider: provider,
		Code:     code,
		State:    state,
	})
	if err != nil {
		log.Error().Err(err).Str("provider", provider).Msg("OAuth callback failed")
		redirectURL := fmt.Sprintf("%s/credentials?oauth=error&error=callback_failed&error_description=%s",
			frontendURL,
			url.QueryEscape(err.Error()))
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	// Redirect to the URL specified in the state (or default credentials page)
	http.Redirect(w, r, result.RedirectURL, http.StatusTemporaryRedirect)
}

// RefreshToken manually refreshes an OAuth token
func (h *OAuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	credentialIDStr := chi.URLParam(r, "credentialID")
	credentialID, err := uuid.Parse(credentialIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid credential ID")
		return
	}

	credential, err := h.oauthSvc.RefreshToken(r.Context(), credentialID)
	if err != nil {
		log.Warn().Err(err).Str("credential_id", credentialIDStr).Msg("Failed to refresh OAuth token")
		dto.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"message":       "Token refreshed successfully",
		"credential_id": credential.ID.String(),
	})
}
