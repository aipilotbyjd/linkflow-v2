package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/pkg/crypto"
	"github.com/linkflow-ai/linkflow/internal/pkg/validator"
)

type AuthHandler struct {
	authSvc    *services.AuthService
	jwtManager *crypto.JWTManager
}

func NewAuthHandler(authSvc *services.AuthService, jwtManager *crypto.JWTManager) *AuthHandler {
	return &AuthHandler{
		authSvc:    authSvc,
		jwtManager: jwtManager,
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
	provider := chi.URLParam(r, "provider")
	_ = provider
	// TODO: Implement OAuth redirect
	dto.ErrorResponse(w, http.StatusNotImplemented, "OAuth not implemented")
}

func (h *AuthHandler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	_ = provider
	// TODO: Implement OAuth callback
	dto.ErrorResponse(w, http.StatusNotImplemented, "OAuth not implemented")
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
