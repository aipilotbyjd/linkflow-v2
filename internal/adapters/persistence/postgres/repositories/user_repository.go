package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/mappers"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/models"
	"github.com/linkflow-ai/linkflow/internal/core/domain/user"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	model := mappers.UserToModel(u)
	return postgres.GetTx(ctx, r.db).Create(model).Error
}

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	model := mappers.UserToModel(u)
	return postgres.GetTx(ctx, r.db).Save(model).Error
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return postgres.GetTx(ctx, r.db).Delete(&models.User{}, "id = ?", id).Error
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	var model models.User
	if err := postgres.GetTx(ctx, r.db).First(&model, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}
	return mappers.UserToDomain(&model), nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	var model models.User
	if err := postgres.GetTx(ctx, r.db).First(&model, "email = ?", email).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}
	return mappers.UserToDomain(&model), nil
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := postgres.GetTx(ctx, r.db).Model(&models.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*user.User, error) {
	var model models.User
	if err := postgres.GetTx(ctx, r.db).First(&model, "username = ?", username).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}
	return mappers.UserToDomain(&model), nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	return postgres.GetTx(ctx, r.db).Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"last_login_at": gorm.Expr("NOW()"),
		}).Error
}

func (r *UserRepository) IncrementFailedLogins(ctx context.Context, userID uuid.UUID) error {
	return postgres.GetTx(ctx, r.db).Model(&models.User{}).
		Where("id = ?", userID).
		UpdateColumn("failed_logins", gorm.Expr("failed_logins + 1")).Error
}

func (r *UserRepository) LockUser(ctx context.Context, userID uuid.UUID, until time.Time) error {
	return postgres.GetTx(ctx, r.db).Model(&models.User{}).
		Where("id = ?", userID).
		Update("locked_until", until).Error
}

func (r *UserRepository) VerifyEmail(ctx context.Context, userID uuid.UUID) error {
	return postgres.GetTx(ctx, r.db).Model(&models.User{}).
		Where("id = ?", userID).
		Update("email_verified", true).Error
}

func (r *UserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	return postgres.GetTx(ctx, r.db).Model(&models.User{}).
		Where("id = ?", userID).
		Update("password_hash", passwordHash).Error
}

func (r *UserRepository) EnableMFA(ctx context.Context, userID uuid.UUID, secret string) error {
	return postgres.GetTx(ctx, r.db).Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"mfa_enabled": true,
			"mfa_secret":  secret,
		}).Error
}

func (r *UserRepository) DisableMFA(ctx context.Context, userID uuid.UUID) error {
	return postgres.GetTx(ctx, r.db).Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"mfa_enabled": false,
			"mfa_secret":  nil,
		}).Error
}
