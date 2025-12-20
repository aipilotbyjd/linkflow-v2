package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"gorm.io/gorm"
)

type CredentialRepository struct {
	*BaseRepository[models.Credential]
}

func NewCredentialRepository(db *gorm.DB) *CredentialRepository {
	return &CredentialRepository{
		BaseRepository: NewBaseRepository[models.Credential](db),
	}
}

func (r *CredentialRepository) FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, opts *ListOptions) ([]models.Credential, int64, error) {
	var credentials []models.Credential
	var total int64

	query := r.DB().WithContext(ctx).Where("workspace_id = ?", workspaceID)
	query.Model(&models.Credential{}).Count(&total)

	if opts != nil {
		query = query.Offset(opts.Offset).Limit(opts.Limit).Order(opts.OrderBy + " " + opts.Order)
	}

	err := query.Find(&credentials).Error
	return credentials, total, err
}

func (r *CredentialRepository) FindByType(ctx context.Context, workspaceID uuid.UUID, credType string) ([]models.Credential, error) {
	var credentials []models.Credential
	err := r.DB().WithContext(ctx).
		Where("workspace_id = ? AND type = ?", workspaceID, credType).
		Order("name ASC").
		Find(&credentials).Error
	return credentials, err
}

func (r *CredentialRepository) CountByWorkspace(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	var count int64
	err := r.DB().WithContext(ctx).Model(&models.Credential{}).
		Where("workspace_id = ?", workspaceID).
		Count(&count).Error
	return count, err
}

func (r *CredentialRepository) UpdateLastUsed(ctx context.Context, credentialID uuid.UUID) error {
	return r.DB().WithContext(ctx).Model(&models.Credential{}).
		Where("id = ?", credentialID).
		Update("last_used_at", time.Now()).Error
}

func (r *CredentialRepository) UpdateData(ctx context.Context, credentialID uuid.UUID, encryptedData string) error {
	return r.DB().WithContext(ctx).Model(&models.Credential{}).
		Where("id = ?", credentialID).
		Update("data", encryptedData).Error
}
