package mappers

import (
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
)

// UserToResponse converts a User model to UserResponse DTO
func UserToResponse(u *models.User) dto.UserResponse {
	return dto.UserResponse{
		ID:            u.ID.String(),
		Email:         u.Email,
		Username:      u.Username,
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		AvatarURL:     u.AvatarURL,
		EmailVerified: u.EmailVerified,
		MFAEnabled:    u.MFAEnabled,
		CreatedAt:     u.CreatedAt.Unix(),
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
