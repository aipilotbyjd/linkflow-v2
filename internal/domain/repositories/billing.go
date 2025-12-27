package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"gorm.io/gorm"
)

type PlanRepository struct {
	*BaseRepository[models.Plan]
}

func NewPlanRepository(db *gorm.DB) *PlanRepository {
	return &PlanRepository{
		BaseRepository: NewBaseRepository[models.Plan](db),
	}
}

func (r *PlanRepository) FindByID(ctx context.Context, id string) (*models.Plan, error) {
	var plan models.Plan
	err := r.DB().WithContext(ctx).First(&plan, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *PlanRepository) FindActive(ctx context.Context) ([]models.Plan, error) {
	var plans []models.Plan
	err := r.DB().WithContext(ctx).
		Where("is_active = ?", true).
		Order("sort_order ASC").
		Find(&plans).Error
	return plans, err
}

// Subscription methods
type SubscriptionRepository struct {
	*BaseRepository[models.Subscription]
}

func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{
		BaseRepository: NewBaseRepository[models.Subscription](db),
	}
}

func (r *SubscriptionRepository) FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (*models.Subscription, error) {
	var subscription models.Subscription
	err := r.DB().WithContext(ctx).
		Preload("Plan").
		Where("workspace_id = ?", workspaceID).
		First(&subscription).Error
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

func (r *SubscriptionRepository) FindByStripeSubscriptionID(ctx context.Context, stripeSubID string) (*models.Subscription, error) {
	var subscription models.Subscription
	err := r.DB().WithContext(ctx).
		Preload("Plan").
		Preload("Workspace").
		Where("stripe_subscription_id = ?", stripeSubID).
		First(&subscription).Error
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

func (r *SubscriptionRepository) UpdateStatus(ctx context.Context, subscriptionID uuid.UUID, status string) error {
	return r.DB().WithContext(ctx).Model(&models.Subscription{}).
		Where("id = ?", subscriptionID).
		Update("status", status).Error
}

func (r *SubscriptionRepository) UpdatePeriod(ctx context.Context, subscriptionID uuid.UUID, start, end time.Time) error {
	return r.DB().WithContext(ctx).Model(&models.Subscription{}).
		Where("id = ?", subscriptionID).
		Updates(map[string]interface{}{
			"current_period_start": start,
			"current_period_end":   end,
		}).Error
}

func (r *SubscriptionRepository) SetCancelAt(ctx context.Context, subscriptionID uuid.UUID, cancelAt *time.Time) error {
	return r.DB().WithContext(ctx).Model(&models.Subscription{}).
		Where("id = ?", subscriptionID).
		Update("cancel_at", cancelAt).Error
}

func (r *SubscriptionRepository) FindExpiringSoon(ctx context.Context, withinDays int) ([]models.Subscription, error) {
	var subscriptions []models.Subscription
	cutoff := time.Now().AddDate(0, 0, withinDays)
	err := r.DB().WithContext(ctx).
		Preload("Workspace").
		Where("status = ? AND current_period_end <= ?", "active", cutoff).
		Find(&subscriptions).Error
	return subscriptions, err
}

// Usage methods
type UsageRepository struct {
	*BaseRepository[models.Usage]
}

func NewUsageRepository(db *gorm.DB) *UsageRepository {
	return &UsageRepository{
		BaseRepository: NewBaseRepository[models.Usage](db),
	}
}

func (r *UsageRepository) FindByWorkspaceAndPeriod(ctx context.Context, workspaceID uuid.UUID, start, end time.Time) (*models.Usage, error) {
	var usage models.Usage
	err := r.DB().WithContext(ctx).
		Where("workspace_id = ? AND period_start = ? AND period_end = ?", workspaceID, start, end).
		First(&usage).Error
	if err != nil {
		return nil, err
	}
	return &usage, nil
}

func (r *UsageRepository) FindCurrentPeriod(ctx context.Context, workspaceID uuid.UUID) (*models.Usage, error) {
	var usage models.Usage
	now := time.Now()
	err := r.DB().WithContext(ctx).
		Where("workspace_id = ? AND period_start <= ? AND period_end >= ?", workspaceID, now, now).
		First(&usage).Error
	if err != nil {
		return nil, err
	}
	return &usage, nil
}

func (r *UsageRepository) IncrementExecutions(ctx context.Context, workspaceID uuid.UUID, periodStart, periodEnd time.Time) error {
	return r.DB().WithContext(ctx).
		Where("workspace_id = ? AND period_start = ? AND period_end = ?", workspaceID, periodStart, periodEnd).
		Update("executions", gorm.Expr("executions + 1")).Error
}

func (r *UsageRepository) UpdateCounts(ctx context.Context, usageID uuid.UUID, workflows, members, credentials int) error {
	return r.DB().WithContext(ctx).Model(&models.Usage{}).
		Where("id = ?", usageID).
		Updates(map[string]interface{}{
			"workflows":   workflows,
			"members":     members,
			"credentials": credentials,
		}).Error
}

// IncrementCredits increments the credits used
func (r *UsageRepository) IncrementCredits(ctx context.Context, workspaceID uuid.UUID, periodStart, periodEnd time.Time, credits int) error {
	return r.DB().WithContext(ctx).Model(&models.Usage{}).
		Where("workspace_id = ? AND period_start = ? AND period_end = ?", workspaceID, periodStart, periodEnd).
		Update("credits_used", gorm.Expr("credits_used + ?", credits)).Error
}

// IncrementOperations increments the operations counter
func (r *UsageRepository) IncrementOperations(ctx context.Context, workspaceID uuid.UUID, periodStart, periodEnd time.Time) error {
	return r.DB().WithContext(ctx).Model(&models.Usage{}).
		Where("workspace_id = ? AND period_start = ? AND period_end = ?", workspaceID, periodStart, periodEnd).
		Update("operations", gorm.Expr("operations + 1")).Error
}

// IncrementExecutionSuccess increments success counter
func (r *UsageRepository) IncrementExecutionSuccess(ctx context.Context, workspaceID uuid.UUID, periodStart, periodEnd time.Time) error {
	return r.DB().WithContext(ctx).Model(&models.Usage{}).
		Where("workspace_id = ? AND period_start = ? AND period_end = ?", workspaceID, periodStart, periodEnd).
		Updates(map[string]interface{}{
			"executions":         gorm.Expr("executions + 1"),
			"executions_success": gorm.Expr("executions_success + 1"),
		}).Error
}

// IncrementExecutionFailed increments failure counter
func (r *UsageRepository) IncrementExecutionFailed(ctx context.Context, workspaceID uuid.UUID, periodStart, periodEnd time.Time) error {
	return r.DB().WithContext(ctx).Model(&models.Usage{}).
		Where("workspace_id = ? AND period_start = ? AND period_end = ?", workspaceID, periodStart, periodEnd).
		Updates(map[string]interface{}{
			"executions":        gorm.Expr("executions + 1"),
			"executions_failed": gorm.Expr("executions_failed + 1"),
		}).Error
}

// IncrementWebhookCalled increments webhook counter
func (r *UsageRepository) IncrementWebhookCalled(ctx context.Context, workspaceID uuid.UUID, periodStart, periodEnd time.Time) error {
	return r.DB().WithContext(ctx).Model(&models.Usage{}).
		Where("workspace_id = ? AND period_start = ? AND period_end = ?", workspaceID, periodStart, periodEnd).
		Update("webhooks_called", gorm.Expr("webhooks_called + 1")).Error
}

// IncrementScheduleTriggered increments schedule counter
func (r *UsageRepository) IncrementScheduleTriggered(ctx context.Context, workspaceID uuid.UUID, periodStart, periodEnd time.Time) error {
	return r.DB().WithContext(ctx).Model(&models.Usage{}).
		Where("workspace_id = ? AND period_start = ? AND period_end = ?", workspaceID, periodStart, periodEnd).
		Update("schedules_triggered", gorm.Expr("schedules_triggered + 1")).Error
}

func (r *UsageRepository) GetOrCreateCurrentPeriod(ctx context.Context, workspaceID uuid.UUID) (*models.Usage, error) {
	now := time.Now()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Second)

	usage, err := r.FindByWorkspaceAndPeriod(ctx, workspaceID, periodStart, periodEnd)
	if err == nil {
		return usage, nil
	}

	if err == gorm.ErrRecordNotFound {
		usage = &models.Usage{
			WorkspaceID: workspaceID,
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
		}
		if err := r.Create(ctx, usage); err != nil {
			return nil, err
		}
		return usage, nil
	}

	return nil, err
}

// Invoice methods
type InvoiceRepository struct {
	*BaseRepository[models.Invoice]
}

func NewInvoiceRepository(db *gorm.DB) *InvoiceRepository {
	return &InvoiceRepository{
		BaseRepository: NewBaseRepository[models.Invoice](db),
	}
}

func (r *InvoiceRepository) FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, opts *ListOptions) ([]models.Invoice, int64, error) {
	var invoices []models.Invoice
	var total int64

	query := r.DB().WithContext(ctx).Where("workspace_id = ?", workspaceID)
	query.Model(&models.Invoice{}).Count(&total)

	if opts != nil {
		query = query.Offset(opts.Offset).Limit(opts.Limit).Order("created_at DESC")
	}

	err := query.Find(&invoices).Error
	return invoices, total, err
}

func (r *InvoiceRepository) FindByStripeInvoiceID(ctx context.Context, stripeInvoiceID string) (*models.Invoice, error) {
	var invoice models.Invoice
	err := r.DB().WithContext(ctx).
		Where("stripe_invoice_id = ?", stripeInvoiceID).
		First(&invoice).Error
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (r *InvoiceRepository) UpdateStatus(ctx context.Context, invoiceID uuid.UUID, status string) error {
	return r.DB().WithContext(ctx).Model(&models.Invoice{}).
		Where("id = ?", invoiceID).
		Update("status", status).Error
}

func (r *InvoiceRepository) MarkPaid(ctx context.Context, invoiceID uuid.UUID, amountPaid int) error {
	now := time.Now()
	return r.DB().WithContext(ctx).Model(&models.Invoice{}).
		Where("id = ?", invoiceID).
		Updates(map[string]interface{}{
			"status":      "paid",
			"amount_paid": amountPaid,
			"paid_at":     now,
		}).Error
}
