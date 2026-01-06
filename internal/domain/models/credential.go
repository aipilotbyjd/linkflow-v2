package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SharingScope defines who can access a credential
type SharingScope string

const (
	SharingScopePrivate   SharingScope = "private"   // Only owner can see/use
	SharingScopeWorkspace SharingScope = "workspace" // All workspace members (default)
	SharingScopeSpecific  SharingScope = "specific"  // Only specific shared users
)

type Credential struct {
	ID                uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID       uuid.UUID      `gorm:"type:uuid;index;not null" json:"workspace_id"`
	CreatedBy         uuid.UUID      `gorm:"type:uuid;not null" json:"created_by"`
	Name              string         `gorm:"size:100;not null" json:"name"`
	Type              string         `gorm:"size:50;not null;index" json:"type"`
	Data              string         `gorm:"type:text;not null" json:"-"` // AES-256-GCM encrypted
	Description       *string        `gorm:"type:text" json:"description,omitempty"`
	Provider          *string        `gorm:"size:50;index" json:"provider,omitempty"`           // OAuth provider (google, slack, etc.)
	ProviderAccountID *string        `gorm:"size:255" json:"provider_account_id,omitempty"`     // Provider's user/account ID
	TokenExpiresAt    *time.Time     `gorm:"index" json:"token_expires_at,omitempty"`           // When OAuth token expires
	SharingScope      SharingScope   `gorm:"size:20;default:'workspace'" json:"sharing_scope"`  // private, workspace, specific
	LastUsedAt        *time.Time     `json:"last_used_at,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`

	Workspace Workspace           `gorm:"foreignKey:WorkspaceID" json:"-"`
	Creator   User                `gorm:"foreignKey:CreatedBy" json:"-"`
	Shares    []CredentialShare   `gorm:"foreignKey:CredentialID" json:"shares,omitempty"`
}

// CredentialShare represents a credential shared with a specific user
type CredentialShare struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CredentialID uuid.UUID      `gorm:"type:uuid;index;not null" json:"credential_id"`
	UserID       uuid.UUID      `gorm:"type:uuid;index;not null" json:"user_id"`
	Permission   string         `gorm:"size:20;default:'use'" json:"permission"` // "use" only for now
	SharedBy     uuid.UUID      `gorm:"type:uuid;not null" json:"shared_by"`
	CreatedAt    time.Time      `json:"created_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	Credential Credential `gorm:"foreignKey:CredentialID" json:"-"`
	User       User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Sharer     User       `gorm:"foreignKey:SharedBy" json:"-"`
}

func (CredentialShare) TableName() string {
	return "credential_shares"
}

func (Credential) TableName() string {
	return "credentials"
}

func (c *Credential) GetWorkspaceID() uuid.UUID {
	return c.WorkspaceID
}

// IsOwner checks if the user is the credential owner
func (c *Credential) IsOwner(userID uuid.UUID) bool {
	return c.CreatedBy == userID
}

// CanUserAccess checks if a user can access this credential
func (c *Credential) CanUserAccess(userID uuid.UUID, isWorkspaceMember bool) bool {
	// Owner always has access
	if c.IsOwner(userID) {
		return true
	}

	switch c.SharingScope {
	case SharingScopePrivate:
		return false
	case SharingScopeWorkspace:
		return isWorkspaceMember
	case SharingScopeSpecific:
		for _, share := range c.Shares {
			if share.UserID == userID && share.DeletedAt.Time.IsZero() {
				return true
			}
		}
		return false
	default:
		// Default to workspace scope for backwards compatibility
		return isWorkspaceMember
	}
}

// CanUserEdit checks if a user can edit/delete this credential
func (c *Credential) CanUserEdit(userID uuid.UUID) bool {
	return c.IsOwner(userID)
}

// CanUserShare checks if a user can share this credential
func (c *Credential) CanUserShare(userID uuid.UUID) bool {
	return c.IsOwner(userID)
}

// CredentialData represents the decrypted credential data structure
type CredentialData struct {
	// Provider info (for OAuth)
	Provider string `json:"provider,omitempty"`

	// API Key
	APIKey string `json:"api_key,omitempty"`

	// OAuth2
	ClientID     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	Scope        string `json:"scope,omitempty"`
	ExpiresAt    string `json:"expires_at,omitempty"` // RFC3339 timestamp

	// Basic Auth
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`

	// Bearer Token
	Token string `json:"token,omitempty"`

	// Database credentials
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	Database string `json:"database,omitempty"`

	// Connection string (for MongoDB etc.)
	ConnectionString string `json:"connectionString,omitempty"`

	// Custom fields
	Custom map[string]string `json:"custom,omitempty"`

	// Generic data map for flexible access
	Data map[string]interface{} `json:"data,omitempty"`
}
