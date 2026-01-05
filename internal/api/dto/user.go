package dto

import "github.com/linkflow-ai/linkflow/internal/domain/models"

// User requests

type UpdateUserRequest struct {
	FirstName               *string     `json:"first_name,omitempty" validate:"omitempty,min=1,max=100"`
	LastName                *string     `json:"last_name,omitempty" validate:"omitempty,min=1,max=100"`
	Username                *string     `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	AvatarURL               *string     `json:"avatar_url,omitempty" validate:"omitempty,url"`
	Phone                   *string     `json:"phone,omitempty" validate:"omitempty,max=20"`
	Bio                     *string     `json:"bio,omitempty" validate:"omitempty,max=500"`
	JobTitle                *string     `json:"job_title,omitempty" validate:"omitempty,max=100"`
	Company                 *string     `json:"company,omitempty" validate:"omitempty,max=100"`
	Timezone                *string     `json:"timezone,omitempty" validate:"omitempty,max=50"`
	Language                *string     `json:"language,omitempty" validate:"omitempty,max=10"`
	DateFormat              *string     `json:"date_format,omitempty" validate:"omitempty,oneof=MM/DD/YYYY DD/MM/YYYY YYYY-MM-DD"`
	TimeFormat              *string     `json:"time_format,omitempty" validate:"omitempty,oneof=12h 24h"`
	Theme                   *string     `json:"theme,omitempty" validate:"omitempty,oneof=light dark system"`
	NotificationPreferences models.JSON `json:"notification_preferences,omitempty"`
}

// User responses

type UserResponse struct {
	ID                      string      `json:"id"`
	Email                   string      `json:"email"`
	Username                *string     `json:"username,omitempty"`
	FirstName               string      `json:"first_name"`
	LastName                string      `json:"last_name"`
	AvatarURL               *string     `json:"avatar_url,omitempty"`
	Phone                   *string     `json:"phone,omitempty"`
	Bio                     *string     `json:"bio,omitempty"`
	JobTitle                *string     `json:"job_title,omitempty"`
	Company                 *string     `json:"company,omitempty"`
	Timezone                string      `json:"timezone"`
	Language                string      `json:"language"`
	DateFormat              string      `json:"date_format"`
	TimeFormat              string      `json:"time_format"`
	Theme                   string      `json:"theme"`
	NotificationPreferences models.JSON `json:"notification_preferences,omitempty"`
	Status                  string      `json:"status"`
	EmailVerified           bool        `json:"email_verified"`
	MFAEnabled              bool        `json:"mfa_enabled"`
	LastLoginAt             *int64      `json:"last_login_at,omitempty"`
	LoginCount              int         `json:"login_count"`
	CreatedAt               int64       `json:"created_at"`
	UpdatedAt               int64       `json:"updated_at"`
}
