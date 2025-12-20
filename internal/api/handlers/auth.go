package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/pkg/crypto"
	"github.com/linkflow-ai/linkflow/internal/pkg/oauth"
	"github.com/linkflow-ai/linkflow/internal/pkg/validator"
)

type AuthHandler struct {
	authSvc      *services.AuthService
	jwtManager   *crypto.JWTManager
	oauthManager *oauth.Manager
	frontendURL  string
}

func NewAuthHandler(authSvc *services.AuthService, jwtManager *crypto.JWTManager) *AuthHandler {
	return &AuthHandler{
		authSvc:    authSvc,
		jwtManager: jwtManager,
	}
}

func NewAuthHandlerWithOAuth(authSvc *services.AuthService, jwtManager *crypto.JWTManager, oauthManager *oauth.Manager, frontendURL string) *AuthHandler {
	return &AuthHandler{
		authSvc:      authSvc,
		jwtManager:   jwtManager,
		oauthManager: oauthManager,
		frontendURL:  frontendURL,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		dto.ValidationErrorResponse(w, err)
		return
	}

	result, err := h.authSvc.Register(r.Context(), services.RegisterInput{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		if err == services.ErrEmailExists {
			dto.ErrorResponse(w, http.StatusConflict, "email already exists")
			return
		}
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to register user")
		return
	}

	dto.Created(w, dto.AuthResponse{
		User: &dto.UserResponse{
			ID:            result.User.ID.String(),
			Email:         result.User.Email,
			FirstName:     result.User.FirstName,
			LastName:      result.User.LastName,
			EmailVerified: result.User.EmailVerified,
			MFAEnabled:    result.User.MFAEnabled,
			CreatedAt:     result.User.CreatedAt.Unix(),
		},
		AccessToken:  result.TokenPair.AccessToken,
		RefreshToken: result.TokenPair.RefreshToken,
		ExpiresAt:    result.TokenPair.ExpiresAt.Unix(),
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		dto.ValidationErrorResponse(w, err)
		return
	}

	result, err := h.authSvc.Login(r.Context(), services.LoginInput{
		Email:     req.Email,
		Password:  req.Password,
		MFACode:   req.MFACode,
		IP:        r.RemoteAddr,
		UserAgent: r.UserAgent(),
	})
	if err != nil {
		switch err {
		case services.ErrInvalidCredentials:
			dto.ErrorResponse(w, http.StatusUnauthorized, "invalid email or password")
		case services.ErrUserLocked:
			dto.ErrorResponse(w, http.StatusForbidden, "account is locked")
		case services.ErrInvalidMFACode:
			dto.ErrorResponse(w, http.StatusUnauthorized, "invalid MFA code")
		default:
			dto.ErrorResponse(w, http.StatusInternalServerError, "login failed")
		}
		return
	}

	if result.RequiresMFA {
		dto.JSON(w, http.StatusOK, dto.MFARequiredResponse{
			RequiresMFA: true,
			Message:     "MFA verification required",
		})
		return
	}

	dto.JSON(w, http.StatusOK, dto.AuthResponse{
		User: &dto.UserResponse{
			ID:            result.User.ID.String(),
			Email:         result.User.Email,
			Username:      result.User.Username,
			FirstName:     result.User.FirstName,
			LastName:      result.User.LastName,
			AvatarURL:     result.User.AvatarURL,
			EmailVerified: result.User.EmailVerified,
			MFAEnabled:    result.User.MFAEnabled,
			CreatedAt:     result.User.CreatedAt.Unix(),
		},
		AccessToken:  result.TokenPair.AccessToken,
		RefreshToken: result.TokenPair.RefreshToken,
		ExpiresAt:    result.TokenPair.ExpiresAt.Unix(),
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		dto.ErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.authSvc.LogoutAll(r.Context(), claims.UserID); err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "logout failed")
		return
	}

	dto.NoContent(w)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tokenPair, err := h.authSvc.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		dto.ErrorResponse(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"expires_at":    tokenPair.ExpiresAt.Unix(),
	})
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req dto.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// TODO: Implement password reset email
	dto.JSON(w, http.StatusOK, map[string]string{
		"message": "If the email exists, a password reset link has been sent",
	})
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req dto.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// TODO: Implement password reset with token
	dto.JSON(w, http.StatusOK, map[string]string{
		"message": "Password has been reset successfully",
	})
}

func (h *AuthHandler) OAuthRedirect(w http.ResponseWriter, r *http.Request) {
	if h.oauthManager == nil {
		dto.ErrorResponse(w, http.StatusNotImplemented, "OAuth not configured")
		return
	}

	provider := chi.URLParam(r, "provider")
	purpose := r.URL.Query().Get("purpose")
	if purpose == "" {
		purpose = "login"
	}

	// Check if user wants to connect account (requires auth)
	var userID uuid.UUID
	var workspaceID uuid.UUID
	if purpose == "connect" {
		claims := middleware.GetUserFromContext(r.Context())
		if claims == nil {
			dto.ErrorResponse(w, http.StatusUnauthorized, "authentication required")
			return
		}
		userID = claims.UserID
		if claims.WorkspaceID != nil {
			workspaceID = *claims.WorkspaceID
		}
	}

	// Determine redirect URI
	redirectURI := fmt.Sprintf("%s/api/v1/auth/oauth/%s/callback", h.frontendURL, provider)

	// Get scopes for provider
	var scopes []string
	switch provider {
	case "google":
		scopes = []string{"openid", "email", "profile"}
	case "github":
		scopes = []string{"user:email", "read:user"}
	case "microsoft":
		scopes = []string{"openid", "email", "profile", "User.Read"}
	default:
		dto.ErrorResponse(w, http.StatusBadRequest, "unsupported provider")
		return
	}

	stateData := &oauth.StateData{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Purpose:     purpose,
	}

	authURL, err := h.oauthManager.GetAuthURL(provider, stateData, redirectURI, scopes)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to generate auth URL")
		return
	}

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (h *AuthHandler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	if h.oauthManager == nil {
		dto.ErrorResponse(w, http.StatusNotImplemented, "OAuth not configured")
		return
	}

	provider := chi.URLParam(r, "provider")
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errParam := r.URL.Query().Get("error")

	// Check for OAuth error
	if errParam != "" {
		errDesc := r.URL.Query().Get("error_description")
		redirectURL := fmt.Sprintf("%s/auth/error?error=%s&description=%s", h.frontendURL, errParam, errDesc)
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	if code == "" || state == "" {
		dto.ErrorResponse(w, http.StatusBadRequest, "missing code or state")
		return
	}

	result, err := h.oauthManager.HandleCallback(r.Context(), provider, code, state)
	if err != nil {
		redirectURL := fmt.Sprintf("%s/auth/error?error=callback_failed&description=%s", h.frontendURL, err.Error())
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	switch result.StateData.Purpose {
	case "login", "signup":
		// Generate JWT tokens
		tokenPair, err := h.jwtManager.GenerateTokenPair(result.User.ID, result.User.Email, nil)
		if err != nil {
			redirectURL := fmt.Sprintf("%s/auth/error?error=token_generation_failed", h.frontendURL)
			http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
			return
		}

		// Redirect to frontend with tokens
		redirectURL := fmt.Sprintf("%s/auth/callback?access_token=%s&refresh_token=%s&is_new=%t",
			h.frontendURL, tokenPair.AccessToken, tokenPair.RefreshToken, result.IsNewUser)
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)

	case "connect":
		// Redirect back to settings page
		redirectURL := fmt.Sprintf("%s/settings/connections?provider=%s&status=connected", h.frontendURL, provider)
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)

	default:
		redirectURL := fmt.Sprintf("%s/auth/callback?provider=%s", h.frontendURL, provider)
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
	}
}

func (h *AuthHandler) SetupMFA(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		dto.ErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	secret, url, err := h.authSvc.SetupMFA(r.Context(), claims.UserID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to setup MFA")
		return
	}

	dto.JSON(w, http.StatusOK, dto.MFASetupResponse{
		Secret: secret,
		QRCode: url,
	})
}

func (h *AuthHandler) VerifyMFA(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		dto.ErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.VerifyMFARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	valid, err := h.authSvc.VerifyMFA(r.Context(), claims.UserID, req.Code)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to verify MFA")
		return
	}

	if !valid {
		dto.ErrorResponse(w, http.StatusUnauthorized, "invalid MFA code")
		return
	}

	dto.JSON(w, http.StatusOK, map[string]bool{"verified": true})
}

func (h *AuthHandler) DisableMFA(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		dto.ErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.VerifyMFARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.authSvc.DisableMFA(r.Context(), claims.UserID, req.Code); err != nil {
		if err == services.ErrInvalidMFACode {
			dto.ErrorResponse(w, http.StatusUnauthorized, "invalid MFA code")
			return
		}
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to disable MFA")
		return
	}

	dto.NoContent(w)
}
