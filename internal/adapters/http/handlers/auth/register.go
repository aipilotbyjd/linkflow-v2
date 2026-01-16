package auth

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	userCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/user"
	"github.com/linkflow-ai/linkflow/internal/core/domain/user"
)

// RegisterRequest represents registration request body
type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
}

// RegisterResponse represents registration response
type RegisterResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt    string       `json:"expires_at"`
	TokenType    string       `json:"token_type"`
}

// UserResponse represents user in responses
type UserResponse struct {
	ID            string  `json:"id"`
	Email         string  `json:"email"`
	FirstName     string  `json:"first_name"`
	LastName      string  `json:"last_name"`
	AvatarURL     *string `json:"avatar_url,omitempty"`
	EmailVerified bool    `json:"email_verified"`
	MFAEnabled    bool    `json:"mfa_enabled"`
	CreatedAt     string  `json:"created_at"`
}

// RegisterHandler handles user registration
type RegisterHandler struct {
	handler *userCmd.RegisterUserHandler
}

// NewRegisterHandler creates a new register handler
func NewRegisterHandler(handler *userCmd.RegisterUserHandler) *RegisterHandler {
	return &RegisterHandler{handler: handler}
}

// Handle handles the registration request
func (h *RegisterHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "invalid request body")
		return
	}

	result, err := h.handler.Handle(r.Context(), userCmd.RegisterUserCommand{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Created(w, RegisterResponse{
		User:         toUserResponse(result.User),
		AccessToken:  result.TokenPair.AccessToken,
		RefreshToken: result.TokenPair.RefreshToken,
		ExpiresAt:    result.TokenPair.ExpiresAt.Format("2006-01-02T15:04:05Z"),
		TokenType:    "Bearer",
	})
}

func toUserResponse(u *user.User) UserResponse {
	return UserResponse{
		ID:            u.ID.String(),
		Email:         u.Email,
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		AvatarURL:     u.AvatarURL,
		EmailVerified: u.EmailVerified,
		MFAEnabled:    u.MFAEnabled,
		CreatedAt:     u.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
