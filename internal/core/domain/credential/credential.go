package credential

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Credential entity (aggregate root)
type Credential struct {
	ID                uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID       uuid.UUID      `gorm:"type:uuid;index;not null" json:"workspace_id"`
	CreatedBy         uuid.UUID      `gorm:"type:uuid;not null" json:"created_by"`
	Name              string         `gorm:"size:100;not null" json:"name"`
	Type              Type           `gorm:"size:50;not null;index" json:"type"`
	Data              string         `gorm:"type:text;not null" json:"-"` // AES-256-GCM encrypted
	Description       *string        `gorm:"type:text" json:"description,omitempty"`
	Provider          *string        `gorm:"size:50;index" json:"provider,omitempty"`
	ProviderAccountID *string        `gorm:"size:255" json:"provider_account_id,omitempty"`
	TokenExpiresAt    *time.Time     `gorm:"index" json:"token_expires_at,omitempty"`
	SharingScope      SharingScope   `gorm:"size:20;default:'workspace'" json:"sharing_scope"`
	LastUsedAt        *time.Time     `json:"last_used_at,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`

	Shares []Share `gorm:"foreignKey:CredentialID" json:"shares,omitempty"`
}

func (Credential) TableName() string {
	return "credentials"
}

// NewCredential creates a new credential with validation
func NewCredential(workspaceID, createdBy uuid.UUID, name string, credType Type, encryptedData string) (*Credential, error) {
	// Validate inputs
	if workspaceID == uuid.Nil {
		return nil, ErrInvalidWorkspaceID
	}
	if createdBy == uuid.Nil {
		return nil, ErrInvalidCreatedBy
	}
	if name == "" {
		return nil, ErrCredentialNameRequired
	}
	if len(name) > 100 {
		return nil, ErrCredentialNameTooLong
	}
	if !credType.IsValid() {
		return nil, ErrInvalidType
	}
	if encryptedData == "" {
		return nil, ErrDataRequired
	}

	return &Credential{
		ID:           uuid.New(),
		WorkspaceID:  workspaceID,
		CreatedBy:    createdBy,
		Name:         name,
		Type:         credType,
		Data:         encryptedData,
		SharingScope: ScopWorkspace,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

// GetWorkspaceID implements the WorkspaceOwned interface
func (c *Credential) GetWorkspaceID() uuid.UUID {
	return c.WorkspaceID
}

// IsOwner checks if the user is the credential owner
func (c *Credential) IsOwner(userID uuid.UUID) bool {
	return c.CreatedBy == userID
}

// CanUserAccess checks if a user can access this credential
func (c *Credential) CanUserAccess(userID uuid.UUID, isWorkspaceMember bool) bool {
	if c.IsOwner(userID) {
		return true
	}

	switch c.SharingScope {
	case ScopePrivate:
		return false
	case ScopWorkspace:
		return isWorkspaceMember
	case ScopeSpecific:
		for _, share := range c.Shares {
			if share.UserID == userID && share.DeletedAt.Time.IsZero() {
				return true
			}
		}
		return false
	default:
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

// Update updates credential details
func (c *Credential) Update(name string, description *string) {
	c.Name = name
	c.Description = description
	c.UpdatedAt = time.Now()
}

// UpdateData updates the encrypted data
func (c *Credential) UpdateData(encryptedData string) {
	c.Data = encryptedData
	c.UpdatedAt = time.Now()
}

// SetProvider sets OAuth provider information
func (c *Credential) SetProvider(provider string, accountID *string) {
	c.Provider = &provider
	c.ProviderAccountID = accountID
	c.UpdatedAt = time.Now()
}

// SetTokenExpiry sets the token expiry time
func (c *Credential) SetTokenExpiry(expiresAt time.Time) {
	c.TokenExpiresAt = &expiresAt
	c.UpdatedAt = time.Now()
}

// SetSharingScope sets the sharing scope
func (c *Credential) SetSharingScope(scope SharingScope) {
	c.SharingScope = scope
	c.UpdatedAt = time.Now()
}

// MarkUsed updates the last used timestamp
func (c *Credential) MarkUsed() {
	now := time.Now()
	c.LastUsedAt = &now
}

// IsExpired checks if the credential token is expired
func (c *Credential) IsExpired() bool {
	if c.TokenExpiresAt == nil {
		return false
	}
	return c.TokenExpiresAt.Before(time.Now())
}

// NeedsRefresh checks if the credential token needs refresh (expires within 5 minutes)
func (c *Credential) NeedsRefresh() bool {
	if c.TokenExpiresAt == nil {
		return false
	}
	return c.TokenExpiresAt.Before(time.Now().Add(5 * time.Minute))
}
