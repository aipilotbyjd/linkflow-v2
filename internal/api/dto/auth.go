package dto

// Auth requests

type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string `json:"last_name" validate:"required,min=1,max=100"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	MFACode  string `json:"mfa_code,omitempty"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

type SetupMFARequest struct {
	Code string `json:"code" validate:"required,len=6"`
}

type VerifyMFARequest struct {
	Code string `json:"code" validate:"required,len=6"`
}

// Auth responses

type AuthResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	ExpiresAt    int64         `json:"expires_at"`
}

type MFARequiredResponse struct {
	RequiresMFA bool   `json:"requires_mfa"`
	Message     string `json:"message"`
}

type MFASetupResponse struct {
	Secret string `json:"secret"`
	QRCode string `json:"qr_code"`
}
