package models

import (
	"time"

	"github.com/google/uuid"
)

// WebhookSignatureAlgorithm constants
const (
	WebhookSigHMACSHA256 = "hmac-sha256"
	WebhookSigHMACSHA1   = "hmac-sha1"
	WebhookSigHMACSHA512 = "hmac-sha512"
)

// WebhookSignatureConfig stores signature verification settings
type WebhookSignatureConfig struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WebhookID       uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"webhook_id"`
	Algorithm       string    `gorm:"size:20;not null;default:hmac-sha256" json:"algorithm"`
	Secret          string    `gorm:"size:255;not null" json:"-"` // Encrypted
	HeaderName      string    `gorm:"size:100;not null;default:X-Signature" json:"header_name"`
	SignaturePrefix *string   `gorm:"size:20" json:"signature_prefix,omitempty"` // e.g., "sha256="
	IsActive        bool      `gorm:"default:true" json:"is_active"`
	FailOnInvalid   bool      `gorm:"default:true" json:"fail_on_invalid"` // Reject invalid signatures
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	Webhook WebhookEndpoint `gorm:"foreignKey:WebhookID" json:"-"`
}

func (WebhookSignatureConfig) TableName() string {
	return "webhook_signature_configs"
}
