package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/core/domain/user"
	"gorm.io/gorm"
)

type OAuthRepository struct {
	db *gorm.DB
}

func NewOAuthRepository(db *gorm.DB) *OAuthRepository {
	return &OAuthRepository{db: db}
}

func (r *OAuthRepository) Create(ctx context.Context, conn *user.OAuthConnection) error {
	return postgres.GetTx(ctx, r.db).Create(conn).Error
}

func (r *OAuthRepository) Update(ctx context.Context, conn *user.OAuthConnection) error {
	return postgres.GetTx(ctx, r.db).Save(conn).Error
}

func (r *OAuthRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return postgres.GetTx(ctx, r.db).Delete(&user.OAuthConnection{}, "id = ?", id).Error
}

func (r *OAuthRepository) FindByID(ctx context.Context, id uuid.UUID) (*user.OAuthConnection, error) {
	var conn user.OAuthConnection
	if err := postgres.GetTx(ctx, r.db).First(&conn, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, user.ErrOAuthConnectionNotFound
		}
		return nil, err
	}
	return &conn, nil
}

func (r *OAuthRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]user.OAuthConnection, error) {
	var connections []user.OAuthConnection
	if err := postgres.GetTx(ctx, r.db).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&connections).Error; err != nil {
		return nil, err
	}
	return connections, nil
}

func (r *OAuthRepository) FindByProviderAndProviderID(ctx context.Context, provider, providerID string) (*user.OAuthConnection, error) {
	var conn user.OAuthConnection
	if err := postgres.GetTx(ctx, r.db).
		Where("provider = ? AND provider_id = ?", provider, providerID).
		First(&conn).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, user.ErrOAuthConnectionNotFound
		}
		return nil, err
	}
	return &conn, nil
}
