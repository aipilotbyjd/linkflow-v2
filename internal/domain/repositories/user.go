package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	*BaseRepository[models.User]
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		BaseRepository: NewBaseRepository[models.User](db),
	}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.DB().WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.DB().WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.DB().WithContext(ctx).Model(&models.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	return r.DB().WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"last_login_at": now,
			"login_count":   gorm.Expr("login_count + 1"),
			"failed_logins": 0,
		}).Error
}

func (r *UserRepository) IncrementFailedLogins(ctx context.Context, userID uuid.UUID) error {
	return r.DB().WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userID).
		Update("failed_logins", gorm.Expr("failed_logins + 1")).Error
}

func (r *UserRepository) LockUser(ctx context.Context, userID uuid.UUID, until time.Time) error {
	return r.DB().WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userID).
		Update("locked_until", until).Error
}

func (r *UserRepository) VerifyEmail(ctx context.Context, userID uuid.UUID) error {
	return r.DB().WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userID).
		Update("email_verified", true).Error
}

func (r *UserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	return r.DB().WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userID).
		Update("password_hash", passwordHash).Error
}

func (r *UserRepository) EnableMFA(ctx context.Context, userID uuid.UUID, secret string) error {
	return r.DB().WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"mfa_enabled": true,
			"mfa_secret":  secret,
		}).Error
}

func (r *UserRepository) DisableMFA(ctx context.Context, userID uuid.UUID) error {
	return r.DB().WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"mfa_enabled": false,
			"mfa_secret":  nil,
		}).Error
}

// Session methods
type SessionRepository struct {
	*BaseRepository[models.Session]
}

func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{
		BaseRepository: NewBaseRepository[models.Session](db),
	}
}

func (r *SessionRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*models.Session, error) {
	var session models.Session
	err := r.DB().WithContext(ctx).
		Preload("User").
		Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", tokenHash, time.Now()).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]models.Session, error) {
	var sessions []models.Session
	err := r.DB().WithContext(ctx).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Order("created_at DESC").
		Find(&sessions).Error
	return sessions, err
}

func (r *SessionRepository) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	now := time.Now()
	return r.DB().WithContext(ctx).Model(&models.Session{}).
		Where("id = ?", sessionID).
		Update("revoked_at", now).Error
}

func (r *SessionRepository) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	return r.DB().WithContext(ctx).Model(&models.Session{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error
}

func (r *SessionRepository) UpdateLastUsed(ctx context.Context, sessionID uuid.UUID) error {
	return r.DB().WithContext(ctx).Model(&models.Session{}).
		Where("id = ?", sessionID).
		Update("last_used_at", time.Now()).Error
}

func (r *SessionRepository) CleanupExpired(ctx context.Context) (int64, error) {
	result := r.DB().WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&models.Session{})
	return result.RowsAffected, result.Error
}

// API Key methods
type APIKeyRepository struct {
	*BaseRepository[models.APIKey]
}

func NewAPIKeyRepository(db *gorm.DB) *APIKeyRepository {
	return &APIKeyRepository{
		BaseRepository: NewBaseRepository[models.APIKey](db),
	}
}

func (r *APIKeyRepository) FindByKeyHash(ctx context.Context, keyHash string) (*models.APIKey, error) {
	var apiKey models.APIKey
	err := r.DB().WithContext(ctx).
		Preload("User").
		Preload("Workspace").
		Where("key_hash = ? AND revoked_at IS NULL", keyHash).
		First(&apiKey).Error
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

func (r *APIKeyRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]models.APIKey, error) {
	var keys []models.APIKey
	err := r.DB().WithContext(ctx).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Order("created_at DESC").
		Find(&keys).Error
	return keys, err
}

func (r *APIKeyRepository) UpdateLastUsed(ctx context.Context, keyID uuid.UUID) error {
	return r.DB().WithContext(ctx).Model(&models.APIKey{}).
		Where("id = ?", keyID).
		Update("last_used_at", time.Now()).Error
}

func (r *APIKeyRepository) Revoke(ctx context.Context, keyID uuid.UUID) error {
	now := time.Now()
	return r.DB().WithContext(ctx).Model(&models.APIKey{}).
		Where("id = ?", keyID).
		Update("revoked_at", now).Error
}
