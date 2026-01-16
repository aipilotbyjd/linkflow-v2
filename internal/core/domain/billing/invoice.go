package billing

import (
	"time"

	"github.com/google/uuid"
)

type InvoiceStatus string

const (
	InvoiceDraft     InvoiceStatus = "draft"
	InvoiceOpen      InvoiceStatus = "open"
	InvoicePaid      InvoiceStatus = "paid"
	InvoiceVoid      InvoiceStatus = "void"
	InvoiceUncollectible InvoiceStatus = "uncollectible"
)

type Invoice struct {
	ID              uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID     uuid.UUID     `gorm:"type:uuid;index;not null" json:"workspace_id"`
	SubscriptionID  uuid.UUID     `gorm:"type:uuid;index;not null" json:"subscription_id"`
	StripeInvoiceID *string       `gorm:"size:100" json:"stripe_invoice_id,omitempty"`
	Number          string        `gorm:"size:50;not null" json:"number"`
	Status          InvoiceStatus `gorm:"size:20;not null;default:draft" json:"status"`
	Currency        string        `gorm:"size:3;not null;default:USD" json:"currency"`
	Subtotal        int64         `gorm:"not null;default:0" json:"subtotal"`
	Tax             int64         `gorm:"default:0" json:"tax"`
	Total           int64         `gorm:"not null;default:0" json:"total"`
	AmountPaid      int64         `gorm:"default:0" json:"amount_paid"`
	AmountDue       int64         `gorm:"default:0" json:"amount_due"`
	PeriodStart     time.Time     `json:"period_start"`
	PeriodEnd       time.Time     `json:"period_end"`
	DueDate         *time.Time    `json:"due_date,omitempty"`
	PaidAt          *time.Time    `json:"paid_at,omitempty"`
	InvoicePDF      *string       `gorm:"type:text" json:"invoice_pdf,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

func (Invoice) TableName() string {
	return "invoices"
}

func NewInvoice(workspaceID, subscriptionID uuid.UUID, number string) *Invoice {
	now := time.Now()
	return &Invoice{
		ID:             uuid.New(),
		WorkspaceID:    workspaceID,
		SubscriptionID: subscriptionID,
		Number:         number,
		Status:         InvoiceDraft,
		Currency:       "USD",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func (i *Invoice) MarkPaid() {
	now := time.Now()
	i.Status = InvoicePaid
	i.PaidAt = &now
	i.AmountPaid = i.Total
	i.AmountDue = 0
	i.UpdatedAt = now
}
