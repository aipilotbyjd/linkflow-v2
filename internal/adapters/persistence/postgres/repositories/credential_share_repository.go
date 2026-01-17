package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/core/domain/credential"
	"gorm.io/gorm"
)

type CredentialShareRepository struct {
	db *gorm.DB
}

func NewCredentialShareRepository(db *gorm.DB) *CredentialShareRepository {
	return &CredentialShareRepository{db: db}
}

func (r *CredentialShareRepository) Create(ctx context.Context, share *credential.Share) error {
	return postgres.GetTx(ctx, r.db).Create(share).Error
}

func (r *CredentialShareRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return postgres.GetTx(ctx, r.db).Delete(&credential.Share{}, "id = ?", id).Error
}

func (r *CredentialShareRepository) FindByID(ctx context.Context, id uuid.UUID) (*credential.Share, error) {
	var share credential.Share
	if err := postgres.GetTx(ctx, r.db).First(&share, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, credential.ErrShareNotFound
		}
		return nil, err
	}
	return &share, nil
}

func (r *CredentialShareRepository) FindByCredentialID(ctx context.Context, credentialID uuid.UUID) ([]credential.Share, error) {
	var shares []credential.Share
	if err := postgres.GetTx(ctx, r.db).
		Where("credential_id = ?", credentialID).
		Find(&shares).Error; err != nil {
		return nil, err
	}
	return shares, nil
}

func (r *CredentialShareRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]credential.Share, error) {
	var shares []credential.Share
	if err := postgres.GetTx(ctx, r.db).
		Where("user_id = ?", userID).
		Preload("Credential").
		Find(&shares).Error; err != nil {
		return nil, err
	}
	return shares, nil
}

func (r *CredentialShareRepository) FindByCredentialAndUser(ctx context.Context, credentialID, userID uuid.UUID) (*credential.Share, error) {
	var share credential.Share
	if err := postgres.GetTx(ctx, r.db).
		Where("credential_id = ? AND user_id = ?", credentialID, userID).
		First(&share).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, credential.ErrShareNotFound
		}
		return nil, err
	}
	return &share, nil
}

func (r *CredentialShareRepository) DeleteByCredentialID(ctx context.Context, credentialID uuid.UUID) error {
	return postgres.GetTx(ctx, r.db).
		Where("credential_id = ?", credentialID).
		Delete(&credential.Share{}).Error
}
