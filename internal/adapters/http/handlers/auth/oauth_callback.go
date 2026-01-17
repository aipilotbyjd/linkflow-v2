package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

type UserCreator interface {
	FindOrCreateOAuthUser(provider, email, name, providerID string) (userID uuid.UUID, isNew bool, err error)
}

type TokenGenerator interface {
	GenerateTokenPair(userID uuid.UUID) (accessToken, refreshToken string, err error)
}

type OAuthCallbackHandler struct {
	oauthManager   OAuthManager
	userCreator    UserCreator
	tokenGenerator TokenGenerator
}

func NewOAuthCallbackHandler(oauthManager OAuthManager, userCreator UserCreator, tokenGenerator TokenGenerator) *OAuthCallbackHandler {
	return &OAuthCallbackHandler{
		oauthManager:   oauthManager,
		userCreator:    userCreator,
		tokenGenerator: tokenGenerator,
	}
}

func (h *OAuthCallbackHandler) Handle(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")

	var req OAuthCallbackRequest

	if r.Method == http.MethodGet {
		req.Code = r.URL.Query().Get("code")
		req.State = r.URL.Query().Get("state")
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
			return
		}
	}

	if req.Code == "" {
		errorMsg := r.URL.Query().Get("error")
		if errorMsg != "" {
			common.Error(w, http.StatusBadRequest, "OAUTH_ERROR", errorMsg)
			return
		}
		common.Error(w, http.StatusBadRequest, "MISSING_CODE", "Authorization code is required")
		return
	}

	userID := uuid.New()
	isNewUser := false

	accessToken := "mock-access-token-" + uuid.New().String()
	refreshToken := "mock-refresh-token-" + uuid.New().String()
	expiresAt := time.Now().Add(15 * time.Minute)

	response := OAuthCallbackResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		User: OAuthUser{
			ID:       userID.String(),
			Email:    "user@example.com",
			Name:     "OAuth User",
			Provider: provider,
		},
		IsNewUser: isNewUser,
	}

	common.Success(w, response)
}
