package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email        string         `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Username     *string        `gorm:"uniqueIndex;size:100" json:"username,omitempty"`
	PasswordHash string         `gorm:"size:255" json:"-"`
	FirstName    string         `gorm:"size:100" json:"first_name"`
	LastName     string         `gorm:"size:100" json:"last_name"`
	AvatarURL    *string        `gorm:"size:500" json:"avatar_url,omitempty"`
	Status       string         `gorm:"size:20;default:active;index" json:"status"`
	EmailVerified bool          `gorm:"default:false" json:"email_verified"`
	MFAEnabled   bool           `gorm:"default:false" json:"mfa_enabled"`
	MFASecret    *string        `gorm:"size:255" json:"-"`
	LastLoginAt  *time.Time     `json:"last_login_at,omitempty"`
	LoginCount   int            `gorm:"default:0" json:"login_count"`
	FailedLogins int            `gorm:"default:0" json:"-"`
	LockedUntil  *time.Time     `json:"-"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Sessions          []Session          `gorm:"foreignKey:UserID" json:"-"`
	APIKeys           []APIKey           `gorm:"foreignKey:UserID" json:"-"`
	OAuthConnections  []OAuthConnection  `gorm:"foreignKey:UserID" json:"-"`
	OwnedWorkspaces   []Workspace        `gorm:"foreignKey:OwnerID" json:"-"`
	WorkspaceMemberships []WorkspaceMember `gorm:"foreignKey:UserID" json:"-"`
}

func (User) TableName() string {
	return "users"
}

type Session struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`
	TokenHash   string     `gorm:"size:255;uniqueIndex;not null" json:"-"`
	RefreshHash *string    `gorm:"size:255" json:"-"`
	IPAddress   *string    `gorm:"size:45" json:"ip_address,omitempty"`
	UserAgent   *string    `gorm:"type:text" json:"user_agent,omitempty"`
	DeviceInfo  JSON       `gorm:"type:jsonb" json:"device_info,omitempty"`
	ExpiresAt   time.Time  `gorm:"not null" json:"expires_at"`
	LastUsedAt  time.Time  `gorm:"default:now()" json:"last_used_at"`
	CreatedAt   time.Time  `json:"created_at"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (Session) TableName() string {
	return "sessions"
}

type APIKey struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`
	WorkspaceID *uuid.UUID `gorm:"type:uuid;index" json:"workspace_id,omitempty"`
	Name        string     `gorm:"size:100;not null" json:"name"`
	KeyPrefix   string     `gorm:"size:10;not null;index" json:"key_prefix"`
	KeyHash     string     `gorm:"size:255;not null" json:"-"`
	Scopes      StringArray `gorm:"type:text[]" json:"scopes"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty"`

	User      User       `gorm:"foreignKey:UserID" json:"-"`
	Workspace *Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
}

func (APIKey) TableName() string {
	return "api_keys"
}

type OAuthConnection struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`
	Provider     string     `gorm:"size:50;not null" json:"provider"`
	ProviderID   string     `gorm:"size:255;not null" json:"provider_id"`
	Email        *string    `gorm:"size:255" json:"email,omitempty"`
	AccessToken  *string    `gorm:"type:text" json:"-"`
	RefreshToken *string    `gorm:"type:text" json:"-"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	ProfileData  JSON       `gorm:"type:jsonb" json:"profile_data,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (OAuthConnection) TableName() string {
	return "oauth_connections"
}

type PasswordResetToken struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`
	Token     string     `gorm:"size:255;uniqueIndex;not null" json:"-"`
	ExpiresAt time.Time  `gorm:"not null" json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}
