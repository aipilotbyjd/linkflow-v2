package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email         string    `gorm:"size:255;uniqueIndex;not null"`
	PasswordHash  string    `gorm:"size:255;not null"`
	FirstName     string    `gorm:"size:100"`
	LastName      string    `gorm:"size:100"`
	AvatarURL     *string   `gorm:"size:500"`
	Phone         *string   `gorm:"size:20"`
	Timezone      string    `gorm:"size:50;default:UTC"`
	Language      string    `gorm:"size:10;default:en"`
	Status        string    `gorm:"size:20;default:active;index"`
	MFAEnabled    bool      `gorm:"default:false"`
	MFASecret     *string   `gorm:"size:100"`
	EmailVerified bool      `gorm:"default:false"`
	LastLoginAt   *time.Time
	LockedUntil   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

func (User) TableName() string {
	return "users"
}

type UserSession struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID       uuid.UUID `gorm:"type:uuid;index;not null"`
	RefreshToken string    `gorm:"size:500;uniqueIndex;not null"`
	UserAgent    *string   `gorm:"size:500"`
	IPAddress    *string   `gorm:"size:45"`
	ExpiresAt    time.Time `gorm:"not null"`
	LastUsedAt   *time.Time
	RevokedAt    *time.Time
	CreatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`

	User User `gorm:"foreignKey:UserID"`
}

func (UserSession) TableName() string {
	return "user_sessions"
}
