package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UsageRepository struct {
	db *gorm.DB
}

func NewUsageRepository(db *gorm.DB) *UsageRepository {
	return &UsageRepository{db: db}
}

func (r *UsageRepository) Create(ctx context.Context, usage *billing.Usage) error {
	return r.db.WithContext(ctx).Create(usage).Error
}

func (r *UsageRepository) Update(ctx context.Context, usage *billing.Usage) error {
	return r.db.WithContext(ctx).Save(usage).Error
}

func (r *UsageRepository) FindByID(ctx context.Context, id uuid.UUID) (*billing.Usage, error) {
	var usage billing.Usage
	if err := r.db.WithContext(ctx).First(&usage, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &usage, nil
}

func (r *UsageRepository) FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (*billing.Usage, error) {
	var usage billing.Usage
	now := time.Now()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	if err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND period_start >= ?", workspaceID, periodStart).
		Order("period_start DESC").
		First(&usage).Error; err != nil {
		return nil, err
	}
	return &usage, nil
}

func (r *UsageRepository) FindByWorkspaceAndPeriod(ctx context.Context, workspaceID uuid.UUID, periodStart, periodEnd time.Time) (*billing.Usage, error) {
	var usage billing.Usage
	if err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND period_start = ? AND period_end = ?", workspaceID, periodStart, periodEnd).
		First(&usage).Error; err != nil {
		return nil, err
	}
	return &usage, nil
}

func (r *UsageRepository) FindAllByWorkspace(ctx context.Context, workspaceID uuid.UUID, limit int) ([]billing.Usage, error) {
	var usages []billing.Usage
	if err := r.db.WithContext(ctx).
		Where("workspace_id = ?", workspaceID).
		Order("period_start DESC").
		Limit(limit).
		Find(&usages).Error; err != nil {
		return nil, err
	}
	return usages, nil
}

// IncrementOperations atomically increments operations count
func (r *UsageRepository) IncrementOperations(ctx context.Context, workspaceID uuid.UUID, count int64) error {
	now := time.Now()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Second)

	// Upsert with atomic increment
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "workspace_id"}, {Name: "period_start"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"operations": gorm.Expr("operations + ?", count),
			"updated_at": time.Now(),
		}),
	}).Create(&billing.Usage{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Operations:  count,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}).Error
}

// IncrementAICredits atomically increments AI credits used
func (r *UsageRepository) IncrementAICredits(ctx context.Context, workspaceID uuid.UUID, count int64) error {
	now := time.Now()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Second)

	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "workspace_id"}, {Name: "period_start"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"ai_credits_used": gorm.Expr("ai_credits_used + ?", count),
			"updated_at":      time.Now(),
		}),
	}).Create(&billing.Usage{
		ID:            uuid.New(),
		WorkspaceID:   workspaceID,
		PeriodStart:   periodStart,
		PeriodEnd:     periodEnd,
		AICreditsUsed: count,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}).Error
}

// GetCurrentPeriodUsage gets or creates usage for current billing period
func (r *UsageRepository) GetCurrentPeriodUsage(ctx context.Context, workspaceID uuid.UUID) (*billing.Usage, error) {
	now := time.Now()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Second)

	var usage billing.Usage
	err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND period_start = ?", workspaceID, periodStart).
		First(&usage).Error

	if err == gorm.ErrRecordNotFound {
		// Create new usage record for this period
		usage = billing.Usage{
			ID:          uuid.New(),
			WorkspaceID: workspaceID,
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := r.db.WithContext(ctx).Create(&usage).Error; err != nil {
			return nil, err
		}
		return &usage, nil
	}

	return &usage, err
}
