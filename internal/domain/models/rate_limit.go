package models

import (
	"time"

	"github.com/google/uuid"
)

// CredentialRateLimit stores per-credential rate limiting config
type CredentialRateLimit struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CredentialID    uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"credential_id"`
	RequestsPerMin  int       `gorm:"default:60" json:"requests_per_min"`
	RequestsPerHour int       `gorm:"default:1000" json:"requests_per_hour"`
	RequestsPerDay  int       `gorm:"default:10000" json:"requests_per_day"`
	BurstLimit      int       `gorm:"default:10" json:"burst_limit"`
	IsActive        bool      `gorm:"default:true" json:"is_active"`
	CurrentMinute   int       `gorm:"default:0" json:"-"`
	CurrentHour     int       `gorm:"default:0" json:"-"`
	CurrentDay      int       `gorm:"default:0" json:"-"`
	LastResetMin    time.Time `json:"-"`
	LastResetHour   time.Time `json:"-"`
	LastResetDay    time.Time `json:"-"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	Credential Credential `gorm:"foreignKey:CredentialID" json:"-"`
}

func (CredentialRateLimit) TableName() string {
	return "credential_rate_limits"
}
