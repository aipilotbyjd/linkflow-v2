package oauth

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
)

type OAuthProvider interface {
	Name() string
	GetAuthURL(state string) string
	ExchangeCode(code string) (Token, error)
	GetUserInfo(token Token) (UserInfo, error)
}

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type UserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type Handler struct {
	providers map[string]OAuthProvider
}

func NewHandler(providers map[string]OAuthProvider) *Handler {
	return &Handler{
		providers: providers,
	}
}

type ProviderInfo struct {
	Name        string `json:"name"`
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Icon        string `json:"icon,omitempty"`
}

func (h *Handler) ListProviders(w http.ResponseWriter, r *http.Request) {
	providers := []ProviderInfo{
		{ID: "google", Name: "google", DisplayName: "Google"},
		{ID: "github", Name: "github", DisplayName: "GitHub"},
		{ID: "slack", Name: "slack", DisplayName: "Slack"},
		{ID: "microsoft", Name: "microsoft", DisplayName: "Microsoft"},
		{ID: "salesforce", Name: "salesforce", DisplayName: "Salesforce"},
		{ID: "hubspot", Name: "hubspot", DisplayName: "HubSpot"},
	}

	common.Success(w, map[string]interface{}{
		"providers": providers,
	})
}

type AuthorizeRequest struct {
	RedirectURL string `json:"redirectUrl"`
	Scopes      string `json:"scopes"`
}

type AuthorizeResponse struct {
	AuthURL string `json:"authUrl"`
	State   string `json:"state"`
}

func (h *Handler) Authorize(w http.ResponseWriter, r *http.Request) {
	providerName := chi.URLParam(r, "provider")

	provider, ok := h.providers[providerName]
	if !ok {
		common.Error(w, http.StatusBadRequest, "INVALID_PROVIDER", "OAuth provider not supported")
		return
	}

	state := uuid.New().String()
	authURL := provider.GetAuthURL(state)

	common.Success(w, AuthorizeResponse{
		AuthURL: authURL,
		State:   state,
	})
}

type CallbackRequest struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

type CallbackResponse struct {
	CredentialID string `json:"credentialId"`
	Provider     string `json:"provider"`
}

func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	providerName := chi.URLParam(r, "provider")
	workspaceID := middleware.GetWorkspaceID(r.Context())

	provider, ok := h.providers[providerName]
	if !ok {
		common.Error(w, http.StatusBadRequest, "INVALID_PROVIDER", "OAuth provider not supported")
		return
	}

	var req CallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		if code == "" || state == "" {
			common.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Missing code or state")
			return
		}
		req.Code = code
		req.State = state
	}

	token, err := provider.ExchangeCode(req.Code)
	if err != nil {
		common.Error(w, http.StatusBadRequest, "TOKEN_EXCHANGE_FAILED", err.Error())
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
