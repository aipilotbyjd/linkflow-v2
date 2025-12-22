package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
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
	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"providers": providers,
	})
}

// AuthorizeRequest represents OAuth authorization request
type AuthorizeRequest struct {
	Provider    string   `json:"provider"`
	Scopes      []string `json:"scopes,omitempty"`
	RedirectURL string   `json:"redirect_url,omitempty"`
}

// Authorize initiates OAuth flow
func (h *OAuthHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	if provider == "" {
		var req AuthorizeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			dto.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
			return
		}
		provider = req.Provider
	}

	userCtx := middleware.GetUserFromContext(r.Context())
	if userCtx == nil {
		dto.ErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "workspace required")
		return
	}

	result, err := h.oauthSvc.GetAuthorizationURL(r.Context(), services.AuthURLInput{
		Provider:    provider,
		UserID:      userCtx.UserID,
		WorkspaceID: wsCtx.WorkspaceID,
	})
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, result)
}

// Callback handles OAuth callback
func (h *OAuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" || state == "" {
		dto.ErrorResponse(w, http.StatusBadRequest, "missing code or state")
		return
	}

	credential, err := h.oauthSvc.HandleCallback(r.Context(), services.CallbackInput{
		Provider: provider,
		Code:     code,
		State:    state,
	})
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"message":       "OAuth connection successful",
		"credential_id": credential.ID,
		"name":          credential.Name,
	})
}

// RefreshToken refreshes an OAuth token
func (h *OAuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	credentialIDStr := chi.URLParam(r, "credentialID")
	credentialID, err := uuid.Parse(credentialIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid credential ID")
		return
	}

	credential, err := h.oauthSvc.RefreshToken(r.Context(), credentialID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"message":       "Token refreshed successfully",
		"credential_id": credential.ID,
	})
}
