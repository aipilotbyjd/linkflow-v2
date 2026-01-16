package credential

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Share represents a credential shared with a specific user
type Share struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CredentialID uuid.UUID      `gorm:"type:uuid;index;not null" json:"credential_id"`
	UserID       uuid.UUID      `gorm:"type:uuid;index;not null" json:"user_id"`
	Permission   Permission     `gorm:"size:20;default:'use'" json:"permission"`
	SharedBy     uuid.UUID      `gorm:"type:uuid;not null" json:"shared_by"`
	CreatedAt    time.Time      `json:"created_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	Credential Credential `gorm:"foreignKey:CredentialID" json:"-"`
}

func (Share) TableName() string {
	return "credential_shares"
}

// NewShare creates a new credential share
func NewShare(credentialID, userID, sharedBy uuid.UUID, permission Permission) *Share {
	return &Share{
		ID:           uuid.New(),
		CredentialID: credentialID,
		UserID:       userID,
		Permission:   permission,
		SharedBy:     sharedBy,
		CreatedAt:    time.Now(),
	}
}

// Permission represents the permission level for a shared credential
type Permission string

const (
	PermissionUse   Permission = "use"
	PermissionView  Permission = "view"
	PermissionEdit  Permission = "edit"
)

func (p Permission) String() string {
	return string(p)
}

func (p Permission) IsValid() bool {
	switch p {
	case PermissionUse, PermissionView, PermissionEdit:
		return true
	default:
		return false
	}
}

// CanUse checks if this permission allows using the credential
func (p Permission) CanUse() bool {
	return p == PermissionUse || p == PermissionEdit
}

// CanView checks if this permission allows viewing the credential
func (p Permission) CanView() bool {
	return true // All permissions allow viewing
}

// CanEdit checks if this permission allows editing the credential
func (p Permission) CanEdit() bool {
	return p == PermissionEdit
}
