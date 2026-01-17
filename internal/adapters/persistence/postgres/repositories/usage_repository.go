package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
	"gorm.io/gorm"
)

type UsageRepository struct {
	db *gorm.DB
}

func NewUsageRepository(db *gorm.DB) *UsageRepository {
	return &UsageRepository{db: db}
}

func (r *UsageRepository) Create(ctx context.Context, usage *billing.Usage) error {
	return postgres.GetTx(ctx, r.db).Create(usage).Error
}

func (r *UsageRepository) Update(ctx context.Context, usage *billing.Usage) error {
	return postgres.GetTx(ctx, r.db).Save(usage).Error
}

func (r *UsageRepository) FindByID(ctx context.Context, id uuid.UUID) (*billing.Usage, error) {
	var usage billing.Usage
	if err := postgres.GetTx(ctx, r.db).First(&usage, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, billing.ErrUsageNotFound
		}
		return nil, err
	}
	return &usage, nil
}

func (r *UsageRepository) FindByWorkspaceAndPeriod(ctx context.Context, workspaceID uuid.UUID, periodStart, periodEnd time.Time) (*billing.Usage, error) {
	var usage billing.Usage
	if err := postgres.GetTx(ctx, r.db).
		Where("workspace_id = ? AND period_start = ? AND period_end = ?", workspaceID, periodStart, periodEnd).
		First(&usage).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, billing.ErrUsageNotFound
		}
		return nil, err
	}
	return &usage, nil
}

func (r *UsageRepository) GetCurrentPeriodUsage(ctx context.Context, workspaceID uuid.UUID) (*billing.Usage, error) {
	var usage billing.Usage
	now := time.Now()
	if err := postgres.GetTx(ctx, r.db).
		Where("workspace_id = ? AND period_start <= ? AND period_end >= ?", workspaceID, now, now).
		First(&usage).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, billing.ErrUsageNotFound
		}
		return nil, err
	}
	return &usage, nil
}

func (r *UsageRepository) IncrementExecutions(ctx context.Context, workspaceID uuid.UUID, count int64) error {
	now := time.Now()
	return postgres.GetTx(ctx, r.db).Model(&billing.Usage{}).
		Where("workspace_id = ? AND period_start <= ? AND period_end >= ?", workspaceID, now, now).
		UpdateColumn("executions", gorm.Expr("executions + ?", count)).
		UpdateColumn("updated_at", now).
		Error
}
