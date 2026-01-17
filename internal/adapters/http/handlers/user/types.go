package user

import (
	"github.com/linkflow-ai/linkflow/internal/core/domain/user"
)

// UserResponse represents user in responses
type UserResponse struct {
	ID            string  `json:"id"`
	Email         string  `json:"email"`
	Username      *string `json:"username,omitempty"`
	FirstName     string  `json:"first_name"`
	LastName      string  `json:"last_name"`
	AvatarURL     *string `json:"avatar_url,omitempty"`
	Phone         *string `json:"phone,omitempty"`
	Bio           *string `json:"bio,omitempty"`
	JobTitle      *string `json:"job_title,omitempty"`
	Company       *string `json:"company,omitempty"`
	Timezone      string  `json:"timezone"`
	Language      string  `json:"language"`
	DateFormat    string  `json:"date_format"`
	TimeFormat    string  `json:"time_format"`
	Theme         string  `json:"theme"`
	EmailVerified bool    `json:"email_verified"`
	MFAEnabled    bool    `json:"mfa_enabled"`
	CreatedAt     string  `json:"created_at"`
}

// UpdateUserRequest represents user update request
type UpdateUserRequest struct {
	FirstName  *string `json:"first_name,omitempty"`
	LastName   *string `json:"last_name,omitempty"`
	AvatarURL  *string `json:"avatar_url,omitempty"`
	Phone      *string `json:"phone,omitempty"`
	Bio        *string `json:"bio,omitempty"`
	JobTitle   *string `json:"job_title,omitempty"`
	Company    *string `json:"company,omitempty"`
	Timezone   *string `json:"timezone,omitempty"`
	Language   *string `json:"language,omitempty"`
	DateFormat *string `json:"date_format,omitempty"`
	TimeFormat *string `json:"time_format,omitempty"`
	Theme      *string `json:"theme,omitempty"`
}

// ToUserResponse converts domain user to response
func ToUserResponse(u *user.User) UserResponse {
	return UserResponse{
		ID:            u.ID.String(),
		Email:         u.Email,
		Username:      u.Username,
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		AvatarURL:     u.AvatarURL,
		Phone:         u.Phone,
		Bio:           u.Bio,
		JobTitle:      u.JobTitle,
		Company:       u.Company,
		Timezone:      u.Timezone,
		Language:      u.Language,
		DateFormat:    u.DateFormat,
		TimeFormat:    u.TimeFormat,
		Theme:         u.Theme,
		EmailVerified: u.EmailVerified,
		MFAEnabled:    u.MFAEnabled,
		CreatedAt:     u.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
