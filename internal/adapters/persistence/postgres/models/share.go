package models

import (
	"time"

	"github.com/google/uuid"
)

// Share represents the database model for shares
type Share struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey"`
	ResourceType    string     `gorm:"type:varchar(50);not null"`
	ResourceID      uuid.UUID  `gorm:"type:uuid;not null;index"`
	ResourceName    string     `gorm:"type:varchar(255);not null"`
	SharedByID      uuid.UUID  `gorm:"type:uuid;not null;index"`
	SharedByEmail   string     `gorm:"type:varchar(255);not null"`
	SharedWithID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	SharedWithEmail string     `gorm:"type:varchar(255);not null"`
	Permission      string     `gorm:"type:varchar(50);not null"`
	Status          string     `gorm:"type:varchar(50);not null;default:'pending'"`
	Message         string     `gorm:"type:text"`
	ExpiresAt       *time.Time `gorm:"type:timestamptz"`
	CreatedAt       time.Time  `gorm:"autoCreateTime"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime"`
}

// TableName returns the table name for Share
func (Share) TableName() string {
	return "shares"
}
