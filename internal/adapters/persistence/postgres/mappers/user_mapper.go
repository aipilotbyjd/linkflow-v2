package mappers

import (
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/models"
	"github.com/linkflow-ai/linkflow/internal/core/domain/user"
)

func UserToModel(u *user.User) *models.User {
	return &models.User{
		ID:            u.ID,
		Email:         u.Email,
		PasswordHash:  u.PasswordHash,
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		AvatarURL:     u.AvatarURL,
		Phone:         u.Phone,
		Timezone:      u.Timezone,
		Language:      u.Language,
		Status:        string(u.Status),
		MFAEnabled:    u.MFAEnabled,
		MFASecret:     u.MFASecret,
		EmailVerified: u.EmailVerified,
		LastLoginAt:   u.LastLoginAt,
		LockedUntil:   u.LockedUntil,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}

func UserToDomain(m *models.User) *user.User {
	return &user.User{
		ID:            m.ID,
		Email:         m.Email,
		PasswordHash:  m.PasswordHash,
		FirstName:     m.FirstName,
		LastName:      m.LastName,
		AvatarURL:     m.AvatarURL,
		Phone:         m.Phone,
		Timezone:      m.Timezone,
		Language:      m.Language,
		Status:        user.Status(m.Status),
		MFAEnabled:    m.MFAEnabled,
		MFASecret:     m.MFASecret,
		EmailVerified: m.EmailVerified,
		LastLoginAt:   m.LastLoginAt,
		LockedUntil:   m.LockedUntil,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}
