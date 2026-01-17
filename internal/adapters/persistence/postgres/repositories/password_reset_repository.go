package repositories

import (
	"context"
	"time"

	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/core/domain/user"
	"gorm.io/gorm"
)

type PasswordResetRepository struct {
	db *gorm.DB
}

func NewPasswordResetRepository(db *gorm.DB) *PasswordResetRepository {
	return &PasswordResetRepository{db: db}
}

func (r *PasswordResetRepository) Create(ctx context.Context, token *user.PasswordResetToken) error {
	return postgres.GetTx(ctx, r.db).Create(token).Error
}

func (r *PasswordResetRepository) FindByToken(ctx context.Context, token string) (*user.PasswordResetToken, error) {
	var resetToken user.PasswordResetToken
	if err := postgres.GetTx(ctx, r.db).
		Where("token = ?", token).
		First(&resetToken).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, user.ErrResetTokenNotFound
		}
		return nil, err
	}
	return &resetToken, nil
}

func (r *PasswordResetRepository) MarkUsed(ctx context.Context, token string) error {
	return postgres.GetTx(ctx, r.db).Model(&user.PasswordResetToken{}).
		Where("token = ?", token).
		Update("used_at", time.Now()).Error
}

func (r *PasswordResetRepository) Delete(ctx context.Context, token string) error {
	return postgres.GetTx(ctx, r.db).
		Where("token = ?", token).
		Delete(&user.PasswordResetToken{}).Error
}

func (r *PasswordResetRepository) CleanupExpired(ctx context.Context) (int64, error) {
	result := postgres.GetTx(ctx, r.db).
		Where("expires_at < ? OR used_at IS NOT NULL", time.Now()).
		Delete(&user.PasswordResetToken{})
	return result.RowsAffected, result.Error
}
