package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
	"gorm.io/gorm"
)

type InvoiceRepository struct {
	db *gorm.DB
}

func NewInvoiceRepository(db *gorm.DB) *InvoiceRepository {
	return &InvoiceRepository{db: db}
}

func (r *InvoiceRepository) Create(ctx context.Context, invoice *billing.Invoice) error {
	return postgres.GetTx(ctx, r.db).Create(invoice).Error
}

func (r *InvoiceRepository) Update(ctx context.Context, invoice *billing.Invoice) error {
	return postgres.GetTx(ctx, r.db).Save(invoice).Error
}

func (r *InvoiceRepository) FindByID(ctx context.Context, id uuid.UUID) (*billing.Invoice, error) {
	var invoice billing.Invoice
	if err := postgres.GetTx(ctx, r.db).First(&invoice, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, billing.ErrInvoiceNotFound
		}
		return nil, err
	}
	return &invoice, nil
}

func (r *InvoiceRepository) FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, limit, offset int) ([]billing.Invoice, int64, error) {
	var invoices []billing.Invoice
	var total int64

	query := postgres.GetTx(ctx, r.db).Model(&billing.Invoice{}).Where("workspace_id = ?", workspaceID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&invoices).Error; err != nil {
		return nil, 0, err
	}

	return invoices, total, nil
}

func (r *InvoiceRepository) FindByStripeInvoiceID(ctx context.Context, stripeID string) (*billing.Invoice, error) {
	var invoice billing.Invoice
	if err := postgres.GetTx(ctx, r.db).
		Where("stripe_invoice_id = ?", stripeID).
		First(&invoice).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, billing.ErrInvoiceNotFound
		}
		return nil, err
	}
	return &invoice, nil
}
