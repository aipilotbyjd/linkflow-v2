package auth

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	userCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/user"
)

// LoginRequest represents login request body
type LoginRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required"`
	MFACode   string `json:"mfa_code,omitempty"`
}

// LoginResponse represents login response
type LoginResponse struct {
	User        UserResponse  `json:"user"`
	AccessToken string        `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt   string        `json:"expires_at"`
	TokenType   string        `json:"token_type"`
	RequiresMFA bool          `json:"requires_mfa,omitempty"`
}

// LoginHandler handles user login
type LoginHandler struct {
	handler *userCmd.LoginUserHandler
}

// NewLoginHandler creates a new login handler
func NewLoginHandler(handler *userCmd.LoginUserHandler) *LoginHandler {
	return &LoginHandler{handler: handler}
}

// Handle handles the login request
func (h *LoginHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "invalid request body")
		return
	}

	// Get client info
	ipAddress := r.Header.Get("X-Forwarded-For")
	if ipAddress == "" {
		ipAddress = r.RemoteAddr
	}
	userAgent := r.UserAgent()

	result, err := h.handler.Handle(r.Context(), userCmd.LoginUserCommand{
		Email:     req.Email,
		Password:  req.Password,
		MFACode:   req.MFACode,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	if result.RequiresMFA {
		common.Success(w, LoginResponse{
			RequiresMFA: true,
		})
		return
	}

	common.Success(w, LoginResponse{
		User:         toUserResponse(result.User),
		AccessToken:  result.TokenPair.AccessToken,
		RefreshToken: result.TokenPair.RefreshToken,
		ExpiresAt:    result.TokenPair.ExpiresAt.Format("2006-01-02T15:04:05Z"),
		TokenType:    "Bearer",
	})
}
