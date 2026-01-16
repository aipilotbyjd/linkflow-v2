package user

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// Repository defines the interface for user persistence
type Repository interface {
	// User operations
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
	IncrementFailedLogins(ctx context.Context, userID uuid.UUID) error
	LockUser(ctx context.Context, userID uuid.UUID, until time.Time) error
	VerifyEmail(ctx context.Context, userID uuid.UUID) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
	EnableMFA(ctx context.Context, userID uuid.UUID, secret string) error
	DisableMFA(ctx context.Context, userID uuid.UUID) error
}

// SessionRepository defines the interface for session persistence
type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	FindByID(ctx context.Context, id uuid.UUID) (*Session, error)
	FindByTokenHash(ctx context.Context, tokenHash string) (*Session, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]Session, error)
	RevokeSession(ctx context.Context, sessionID uuid.UUID) error
	RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	UpdateLastUsed(ctx context.Context, sessionID uuid.UUID) error
	CleanupExpired(ctx context.Context) (int64, error)
}

// APIKeyRepository defines the interface for API key persistence
type APIKeyRepository interface {
	Create(ctx context.Context, key *APIKey) error
	FindByID(ctx context.Context, id uuid.UUID) (*APIKey, error)
	FindByKeyHash(ctx context.Context, keyHash string) (*APIKey, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]APIKey, error)
	FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]APIKey, error)
	UpdateLastUsed(ctx context.Context, keyID uuid.UUID) error
	Revoke(ctx context.Context, keyID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// OAuthRepository defines the interface for OAuth connection persistence
type OAuthRepository interface {
	Create(ctx context.Context, conn *OAuthConnection) error
	Update(ctx context.Context, conn *OAuthConnection) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*OAuthConnection, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]OAuthConnection, error)
	FindByProviderAndProviderID(ctx context.Context, provider, providerID string) (*OAuthConnection, error)
}

// PasswordResetRepository defines the interface for password reset token persistence
type PasswordResetRepository interface {
	Create(ctx context.Context, token *PasswordResetToken) error
	FindByToken(ctx context.Context, token string) (*PasswordResetToken, error)
	MarkUsed(ctx context.Context, token string) error
	Delete(ctx context.Context, token string) error
	CleanupExpired(ctx context.Context) (int64, error)
}

// ListOptions for user queries
type ListOptions struct {
	*types.ListOptions
	Status *Status
	Search string
}

// UserList repository with listing support
type UserListRepository interface {
	Repository
	FindAll(ctx context.Context, opts *ListOptions) ([]User, int64, error)
}
