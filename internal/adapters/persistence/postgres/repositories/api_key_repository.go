package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/core/domain/user"
	"gorm.io/gorm"
)

type APIKeyRepository struct {
	db *gorm.DB
}

func NewAPIKeyRepository(db *gorm.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

func (r *APIKeyRepository) Create(ctx context.Context, key *user.APIKey) error {
	return postgres.GetTx(ctx, r.db).Create(key).Error
}

func (r *APIKeyRepository) FindByID(ctx context.Context, id uuid.UUID) (*user.APIKey, error) {
	var key user.APIKey
	if err := postgres.GetTx(ctx, r.db).First(&key, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, user.ErrAPIKeyNotFound
		}
		return nil, err
	}
	return &key, nil
}

func (r *APIKeyRepository) FindByKeyHash(ctx context.Context, keyHash string) (*user.APIKey, error) {
	var key user.APIKey
	if err := postgres.GetTx(ctx, r.db).First(&key, "key_hash = ?", keyHash).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, user.ErrAPIKeyNotFound
		}
		return nil, err
	}
	return &key, nil
}

func (r *APIKeyRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]user.APIKey, error) {
	var keys []user.APIKey
	if err := postgres.GetTx(ctx, r.db).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Order("created_at DESC").
		Find(&keys).Error; err != nil {
		return nil, err
	}
	return keys, nil
}

func (r *APIKeyRepository) FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]user.APIKey, error) {
	var keys []user.APIKey
	if err := postgres.GetTx(ctx, r.db).
		Where("workspace_id = ?", workspaceID).
		Order("created_at DESC").
		Find(&keys).Error; err != nil {
		return nil, err
	}
	return keys, nil
}

func (r *APIKeyRepository) UpdateLastUsed(ctx context.Context, keyID uuid.UUID) error {
	return postgres.GetTx(ctx, r.db).Model(&user.APIKey{}).
		Where("id = ?", keyID).
		UpdateColumn("last_used_at", time.Now()).
		Error
}

func (r *APIKeyRepository) Revoke(ctx context.Context, keyID uuid.UUID) error {
	now := time.Now()
	return postgres.GetTx(ctx, r.db).Model(&user.APIKey{}).
		Where("id = ?", keyID).
		Updates(map[string]interface{}{
			"revoked_at": now,
			"updated_at": now,
		}).Error
}

func (r *APIKeyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return postgres.GetTx(ctx, r.db).Delete(&user.APIKey{}, "id = ?", id).Error
}
