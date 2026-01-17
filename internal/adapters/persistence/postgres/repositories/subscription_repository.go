package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
	"gorm.io/gorm"
)

type SubscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Create(ctx context.Context, subscription *billing.Subscription) error {
	return postgres.GetTx(ctx, r.db).Create(subscription).Error
}

func (r *SubscriptionRepository) Update(ctx context.Context, subscription *billing.Subscription) error {
	return postgres.GetTx(ctx, r.db).Save(subscription).Error
}

func (r *SubscriptionRepository) FindByID(ctx context.Context, id uuid.UUID) (*billing.Subscription, error) {
	var subscription billing.Subscription
	if err := postgres.GetTx(ctx, r.db).First(&subscription, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, billing.ErrSubscriptionNotFound
		}
		return nil, err
	}
	return &subscription, nil
}

func (r *SubscriptionRepository) FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (*billing.Subscription, error) {
	var subscription billing.Subscription
	if err := postgres.GetTx(ctx, r.db).
		Where("workspace_id = ?", workspaceID).
		Order("created_at DESC").
		First(&subscription).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, billing.ErrSubscriptionNotFound
		}
		return nil, err
	}
	return &subscription, nil
}

func (r *SubscriptionRepository) FindByStripeSubscriptionID(ctx context.Context, stripeID string) (*billing.Subscription, error) {
	var subscription billing.Subscription
	if err := postgres.GetTx(ctx, r.db).
		Where("stripe_subscription_id = ?", stripeID).
		First(&subscription).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, billing.ErrSubscriptionNotFound
		}
		return nil, err
	}
	return &subscription, nil
}
