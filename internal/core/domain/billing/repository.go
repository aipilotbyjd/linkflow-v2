package billing

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, subscription *Subscription) error
	Update(ctx context.Context, subscription *Subscription) error
	FindByID(ctx context.Context, id uuid.UUID) (*Subscription, error)
	FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (*Subscription, error)
	FindByStripeSubscriptionID(ctx context.Context, stripeID string) (*Subscription, error)
}

type UsageRepository interface {
	Create(ctx context.Context, usage *Usage) error
	Update(ctx context.Context, usage *Usage) error
	FindByID(ctx context.Context, id uuid.UUID) (*Usage, error)
	FindByWorkspaceAndPeriod(ctx context.Context, workspaceID uuid.UUID, periodStart, periodEnd time.Time) (*Usage, error)
	GetCurrentPeriodUsage(ctx context.Context, workspaceID uuid.UUID) (*Usage, error)
	IncrementExecutions(ctx context.Context, workspaceID uuid.UUID, count int64) error
}

type InvoiceRepository interface {
	Create(ctx context.Context, invoice *Invoice) error
	Update(ctx context.Context, invoice *Invoice) error
	FindByID(ctx context.Context, id uuid.UUID) (*Invoice, error)
	FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, limit, offset int) ([]Invoice, int64, error)
	FindByStripeInvoiceID(ctx context.Context, stripeID string) (*Invoice, error)
}
