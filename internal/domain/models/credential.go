package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Credential struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID uuid.UUID      `gorm:"type:uuid;index;not null" json:"workspace_id"`
	CreatedBy   uuid.UUID      `gorm:"type:uuid;not null" json:"created_by"`
	Name        string         `gorm:"size:100;not null" json:"name"`
	Type        string         `gorm:"size:50;not null;index" json:"type"`
	Data        string         `gorm:"type:text;not null" json:"-"` // AES-256-GCM encrypted
	Description *string        `gorm:"type:text" json:"description,omitempty"`
	LastUsedAt  *time.Time     `json:"last_used_at,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
	Creator   User      `gorm:"foreignKey:CreatedBy" json:"-"`
}

func (Credential) TableName() string {
	return "credentials"
}

// CredentialData represents the decrypted credential data structure
type CredentialData struct {
	// API Key
	APIKey string `json:"api_key,omitempty"`

	// OAuth2
	ClientID     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	Scope        string `json:"scope,omitempty"`

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
