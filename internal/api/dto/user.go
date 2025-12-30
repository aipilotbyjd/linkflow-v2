package dto

// User requests

type UpdateUserRequest struct {
	FirstName *string `json:"first_name,omitempty" validate:"omitempty,min=1,max=100"`
	LastName  *string `json:"last_name,omitempty" validate:"omitempty,min=1,max=100"`
	Username  *string `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	AvatarURL *string `json:"avatar_url,omitempty" validate:"omitempty,url"`
}

// User responses

type UserResponse struct {
	ID            string  `json:"id"`
	Email         string  `json:"email"`
	Username      *string `json:"username,omitempty"`
	FirstName     string  `json:"first_name"`
	LastName      string  `json:"last_name"`
	AvatarURL     *string `json:"avatar_url,omitempty"`
	EmailVerified bool    `json:"email_verified"`
	MFAEnabled    bool    `json:"mfa_enabled"`
	CreatedAt     int64   `json:"created_at"`
}
