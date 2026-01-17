package auth

import (
	"time"

	"github.com/linkflow-ai/linkflow/internal/core/domain/user"
)

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

// LoginRequest represents login request body
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	MFACode  string `json:"mfa_code,omitempty"`
}

// LoginResponse represents login response
type LoginResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt    string       `json:"expires_at"`
	TokenType    string       `json:"token_type"`
	RequiresMFA  bool         `json:"requires_mfa,omitempty"`
}

// RefreshRequest represents refresh request body
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshResponse represents refresh response
type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    string `json:"expires_at"`
	TokenType    string `json:"token_type"`
}

// ForgotPasswordRequest represents forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPasswordRequest represents reset password request
type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// SetupMFAResponse represents MFA setup response
type SetupMFAResponse struct {
	Secret    string `json:"secret"`
	QRCodeURL string `json:"qr_code_url"`
}

// VerifyMFARequest represents MFA verification request
type VerifyMFARequest struct {
	Code string `json:"code" validate:"required,len=6"`
}

// DisableMFARequest represents MFA disable request
type DisableMFARequest struct {
	Code     string `json:"code" validate:"required,len=6"`
	Password string `json:"password" validate:"required"`
}

// OAuthCallbackRequest represents OAuth callback request
type OAuthCallbackRequest struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

// OAuthCallbackResponse represents OAuth callback response
type OAuthCallbackResponse struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
	User         OAuthUser `json:"user"`
	IsNewUser    bool      `json:"isNewUser"`
}

// OAuthUser represents OAuth user info
type OAuthUser struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
}

// ToUserResponse converts domain user to response
func ToUserResponse(u *user.User) UserResponse {
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
