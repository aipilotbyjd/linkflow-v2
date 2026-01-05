package mappers

import (
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
)

// UserToResponse converts a User model to UserResponse DTO
func UserToResponse(u *models.User) dto.UserResponse {
	var lastLoginAt *int64
	if u.LastLoginAt != nil {
		ts := u.LastLoginAt.Unix()
		lastLoginAt = &ts
	}

	return dto.UserResponse{
		ID:                      u.ID.String(),
		Email:                   u.Email,
		Username:                u.Username,
		FirstName:               u.FirstName,
		LastName:                u.LastName,
		AvatarURL:               u.AvatarURL,
		Phone:                   u.Phone,
		Bio:                     u.Bio,
		JobTitle:                u.JobTitle,
		Company:                 u.Company,
		Timezone:                u.Timezone,
		Language:                u.Language,
		DateFormat:              u.DateFormat,
		TimeFormat:              u.TimeFormat,
		Theme:                   u.Theme,
		NotificationPreferences: u.NotificationPreferences,
		Status:                  u.Status,
		EmailVerified:           u.EmailVerified,
		MFAEnabled:              u.MFAEnabled,
		LastLoginAt:             lastLoginAt,
		LoginCount:              u.LoginCount,
		CreatedAt:               u.CreatedAt.Unix(),
		UpdatedAt:               u.UpdatedAt.Unix(),
	}
}

// UsersToResponse converts a slice of User models to UserResponse DTOs
func UsersToResponse(users []models.User) []dto.UserResponse {
	result := make([]dto.UserResponse, len(users))
	for i := range users {
		result[i] = UserToResponse(&users[i])
	}
	return result
}
