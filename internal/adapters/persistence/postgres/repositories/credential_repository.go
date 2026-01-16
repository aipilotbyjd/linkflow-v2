package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/core/domain/credential"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
	"gorm.io/gorm"
)

type CredentialRepository struct {
	db *gorm.DB
}

func NewCredentialRepository(db *gorm.DB) *CredentialRepository {
	return &CredentialRepository{db: db}
}

func (r *CredentialRepository) Create(ctx context.Context, c *credential.Credential) error {
	return postgres.GetTx(ctx, r.db).Create(c).Error
}

func (r *CredentialRepository) Update(ctx context.Context, c *credential.Credential) error {
	return postgres.GetTx(ctx, r.db).Save(c).Error
}

func (r *CredentialRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return postgres.GetTx(ctx, r.db).Delete(&credential.Credential{}, "id = ?", id).Error
}

func (r *CredentialRepository) FindByID(ctx context.Context, id uuid.UUID) (*credential.Credential, error) {
	var c credential.Credential
	if err := postgres.GetTx(ctx, r.db).First(&c, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, credential.ErrCredentialNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *CredentialRepository) FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, opts *credential.ListOptions) ([]credential.Credential, int64, error) {
	var credentials []credential.Credential
	var total int64

	query := postgres.GetTx(ctx, r.db).Model(&credential.Credential{}).Where("workspace_id = ?", workspaceID)

	if opts != nil {
		if opts.Type != nil {
			query = query.Where("type = ?", *opts.Type)
		}
		if opts.Provider != nil {
			query = query.Where("provider = ?", *opts.Provider)
		}
		if opts.Search != "" {
			query = query.Where("name ILIKE ?", "%"+opts.Search+"%")
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if opts != nil && opts.ListOptions != nil {
		query = query.Offset(opts.Offset).Limit(opts.Limit)
	}

	query = query.Order("created_at DESC")

	if err := query.Find(&credentials).Error; err != nil {
		return nil, 0, err
	}

	return credentials, total, nil
}

func (r *CredentialRepository) FindByName(ctx context.Context, workspaceID uuid.UUID, name string) (*credential.Credential, error) {
	var c credential.Credential
	if err := postgres.GetTx(ctx, r.db).
		Where("workspace_id = ? AND name = ?", workspaceID, name).
		First(&c).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, credential.ErrCredentialNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *CredentialRepository) FindByType(ctx context.Context, workspaceID uuid.UUID, credType credential.Type) ([]credential.Credential, error) {
	var credentials []credential.Credential
	if err := postgres.GetTx(ctx, r.db).
		Where("workspace_id = ? AND type = ?", workspaceID, credType).
		Find(&credentials).Error; err != nil {
		return nil, err
	}
	return credentials, nil
}

func (r *CredentialRepository) ExistsByName(ctx context.Context, workspaceID uuid.UUID, name string) (bool, error) {
	var count int64
	err := postgres.GetTx(ctx, r.db).Model(&credential.Credential{}).
		Where("workspace_id = ? AND name = ?", workspaceID, name).
		Count(&count).Error
	return count > 0, err
}

func (r *CredentialRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	return postgres.GetTx(ctx, r.db).Model(&credential.Credential{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_used_at": gorm.Expr("NOW()"),
			"use_count":    gorm.Expr("use_count + 1"),
		}).Error
}

func (r *CredentialRepository) CountByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	var count int64
	err := postgres.GetTx(ctx, r.db).Model(&credential.Credential{}).Where("workspace_id = ?", workspaceID).Count(&count).Error
	return count, err
}

func (r *CredentialRepository) FindSharedWithUser(ctx context.Context, userID uuid.UUID, opts *types.ListOptions) ([]credential.Credential, int64, error) {
	var credentials []credential.Credential
	var total int64

	query := postgres.GetTx(ctx, r.db).Model(&credential.Credential{}).
		Joins("JOIN credential_shares ON credential_shares.credential_id = credentials.id").
		Where("credential_shares.user_id = ? AND credential_shares.deleted_at IS NULL", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if opts != nil {
		query = query.Offset(opts.Offset).Limit(opts.Limit)
	}

	if err := query.Find(&credentials).Error; err != nil {
		return nil, 0, err
	}

	return credentials, total, nil
}

func (r *CredentialRepository) FindByIDWithShares(ctx context.Context, id uuid.UUID) (*credential.Credential, error) {
	var c credential.Credential
	if err := postgres.GetTx(ctx, r.db).
		Preload("Shares").
		First(&c, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, credential.ErrCredentialNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *CredentialRepository) FindByProvider(ctx context.Context, workspaceID uuid.UUID, provider string) ([]credential.Credential, error) {
	var credentials []credential.Credential
	if err := postgres.GetTx(ctx, r.db).
		Where("workspace_id = ? AND provider = ?", workspaceID, provider).
		Find(&credentials).Error; err != nil {
		return nil, err
	}
	return credentials, nil
}

func (r *CredentialRepository) FindExpiring(ctx context.Context, withinDuration string) ([]credential.Credential, error) {
	var credentials []credential.Credential
	if err := postgres.GetTx(ctx, r.db).
		Where("expires_at IS NOT NULL AND expires_at < NOW() + ?::interval", withinDuration).
		Find(&credentials).Error; err != nil {
		return nil, err
	}
	return credentials, nil
}
