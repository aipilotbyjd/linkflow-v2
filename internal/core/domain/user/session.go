package user

import (
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// Session represents a user session
type Session struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`
	TokenHash   string     `gorm:"size:255;uniqueIndex;not null" json:"-"`
	RefreshHash *string    `gorm:"size:255" json:"-"`
	IPAddress   *string    `gorm:"size:45" json:"ip_address,omitempty"`
	UserAgent   *string    `gorm:"type:text" json:"user_agent,omitempty"`
	DeviceInfo  types.JSON `gorm:"type:jsonb" json:"device_info,omitempty"`
	ExpiresAt   time.Time  `gorm:"not null" json:"expires_at"`
	LastUsedAt  time.Time  `gorm:"default:now()" json:"last_used_at"`
	CreatedAt   time.Time  `json:"created_at"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (Session) TableName() string {
	return "sessions"
}

// NewSession creates a new session
func NewSession(userID uuid.UUID, tokenHash string, expiresAt time.Time) *Session {
	return &Session{
		ID:         uuid.New(),
		UserID:     userID,
		TokenHash:  tokenHash,
		ExpiresAt:  expiresAt,
		LastUsedAt: time.Now(),
		CreatedAt:  time.Now(),
	}
}

// WithRefreshToken adds a refresh token hash
func (s *Session) WithRefreshToken(refreshHash string) *Session {
	s.RefreshHash = &refreshHash
	return s
}

// WithDeviceInfo adds device information
func (s *Session) WithDeviceInfo(ipAddress, userAgent string, deviceInfo types.JSON) *Session {
	s.IPAddress = &ipAddress
	s.UserAgent = &userAgent
	s.DeviceInfo = deviceInfo
	return s
}

// IsExpired checks if the session is expired
func (s *Session) IsExpired() bool {
	return s.ExpiresAt.Before(time.Now())
}

// IsRevoked checks if the session is revoked
func (s *Session) IsRevoked() bool {
	return s.RevokedAt != nil
}

// IsValid checks if the session is valid (not expired and not revoked)
func (s *Session) IsValid() bool {
	return !s.IsExpired() && !s.IsRevoked()
}

// Revoke revokes the session
func (s *Session) Revoke() {
	now := time.Now()
	s.RevokedAt = &now
}

// UpdateLastUsed updates the last used timestamp
func (s *Session) UpdateLastUsed() {
	s.LastUsedAt = time.Now()
}

// Extend extends the session expiration
func (s *Session) Extend(duration time.Duration) {
	s.ExpiresAt = time.Now().Add(duration)
}

// APIKey represents an API key for programmatic access
type APIKey struct {
	ID          uuid.UUID         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID         `gorm:"type:uuid;index;not null" json:"user_id"`
	WorkspaceID *uuid.UUID        `gorm:"type:uuid;index" json:"workspace_id,omitempty"`
	Name        string            `gorm:"size:100;not null" json:"name"`
	KeyPrefix   string            `gorm:"size:10;not null;index" json:"key_prefix"`
	KeyHash     string            `gorm:"size:255;not null" json:"-"`
	Scopes      types.StringArray `gorm:"type:text[]" json:"scopes"`
	LastUsedAt  *time.Time        `json:"last_used_at,omitempty"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	RevokedAt   *time.Time        `json:"revoked_at,omitempty"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (APIKey) TableName() string {
	return "api_keys"
}

// NewAPIKey creates a new API key
func NewAPIKey(userID uuid.UUID, name, keyPrefix, keyHash string, scopes []string) *APIKey {
	return &APIKey{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      name,
		KeyPrefix: keyPrefix,
		KeyHash:   keyHash,
		Scopes:    scopes,
		CreatedAt: time.Now(),
	}
}

// WithWorkspace associates the API key with a workspace
func (k *APIKey) WithWorkspace(workspaceID uuid.UUID) *APIKey {
	k.WorkspaceID = &workspaceID
	return k
}

// WithExpiration sets the expiration time
func (k *APIKey) WithExpiration(expiresAt time.Time) *APIKey {
	k.ExpiresAt = &expiresAt
	return k
}

// IsExpired checks if the API key is expired
func (k *APIKey) IsExpired() bool {
	if k.ExpiresAt == nil {
		return false
	}
	return k.ExpiresAt.Before(time.Now())
}

// IsRevoked checks if the API key is revoked
func (k *APIKey) IsRevoked() bool {
	return k.RevokedAt != nil
}

// IsValid checks if the API key is valid
func (k *APIKey) IsValid() bool {
	return !k.IsExpired() && !k.IsRevoked()
}

// Revoke revokes the API key
func (k *APIKey) Revoke() {
	now := time.Now()
	k.RevokedAt = &now
}

// UpdateLastUsed updates the last used timestamp
func (k *APIKey) UpdateLastUsed() {
	now := time.Now()
	k.LastUsedAt = &now
}

// HasScope checks if the API key has the given scope
func (k *APIKey) HasScope(scope string) bool {
	for _, s := range k.Scopes {
		if s == scope || s == "*" {
			return true
		}
	}
	return false
}

// OAuthConnection represents an OAuth provider connection
type OAuthConnection struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`
	Provider     string     `gorm:"size:50;not null" json:"provider"`
	ProviderID   string     `gorm:"size:255;not null" json:"provider_id"`
	Email        *string    `gorm:"size:255" json:"email,omitempty"`
	AccessToken  *string    `gorm:"type:text" json:"-"`
	RefreshToken *string    `gorm:"type:text" json:"-"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	ProfileData  types.JSON `gorm:"type:jsonb" json:"profile_data,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (OAuthConnection) TableName() string {
	return "oauth_connections"
}

// PasswordResetToken represents a password reset token
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

// IsExpired checks if the token is expired
func (t *PasswordResetToken) IsExpired() bool {
	return t.ExpiresAt.Before(time.Now())
}

// IsUsed checks if the token has been used
func (t *PasswordResetToken) IsUsed() bool {
	return t.UsedAt != nil
}

// IsValid checks if the token is valid
func (t *PasswordResetToken) IsValid() bool {
	return !t.IsExpired() && !t.IsUsed()
}

// MarkUsed marks the token as used
func (t *PasswordResetToken) MarkUsed() {
	now := time.Now()
	t.UsedAt = &now
}
