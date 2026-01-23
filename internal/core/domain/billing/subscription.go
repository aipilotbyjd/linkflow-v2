package billing

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SubscriptionStatus string

const (
	StatusActive   SubscriptionStatus = "active"
	StatusCanceled SubscriptionStatus = "canceled"
	StatusPastDue  SubscriptionStatus = "past_due"
	StatusTrialing SubscriptionStatus = "trialing"
	StatusInactive SubscriptionStatus = "inactive"
)

type Subscription struct {
	ID                   uuid.UUID          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID          uuid.UUID          `gorm:"type:uuid;index;not null" json:"workspace_id"`
	PlanID               string             `gorm:"size:50;not null" json:"plan_id"`
	Status               SubscriptionStatus `gorm:"size:20;not null;default:active" json:"status"`
	StripeCustomerID     *string            `gorm:"size:100" json:"stripe_customer_id,omitempty"`
	StripeSubscriptionID *string            `gorm:"size:100" json:"stripe_subscription_id,omitempty"`
	CurrentPeriodStart   time.Time          `json:"current_period_start"`
	CurrentPeriodEnd     time.Time          `json:"current_period_end"`
	CanceledAt           *time.Time         `json:"canceled_at,omitempty"`
	CancelAtPeriodEnd    bool               `gorm:"default:false" json:"cancel_at_period_end"`
	TrialStart           *time.Time         `json:"trial_start,omitempty"`
	TrialEnd             *time.Time         `json:"trial_end,omitempty"`
	CreatedAt            time.Time          `json:"created_at"`
	UpdatedAt            time.Time          `json:"updated_at"`
	DeletedAt            gorm.DeletedAt     `gorm:"index" json:"-"`
}

func (Subscription) TableName() string {
	return "subscriptions"
}

func NewSubscription(workspaceID uuid.UUID, planID string) (*Subscription, error) {
	if workspaceID == uuid.Nil {
		return nil, ErrInvalidWorkspaceID
	}
	if planID == "" {
		return nil, ErrPlanIDRequired
	}

	now := time.Now()
	return &Subscription{
		ID:                 uuid.New(),
		WorkspaceID:        workspaceID,
		PlanID:             planID,
		Status:             StatusActive,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   now.AddDate(0, 1, 0),
		CreatedAt:          now,
		UpdatedAt:          now,
	}, nil
}

func (s *Subscription) Cancel() {
	now := time.Now()
	s.CanceledAt = &now
	s.Status = StatusCanceled
	s.UpdatedAt = now
}

func (s *Subscription) Reactivate() {
	s.CanceledAt = nil
	s.Status = StatusActive
	s.CancelAtPeriodEnd = false
	s.UpdatedAt = time.Now()
}

func (s *Subscription) IsActive() bool {
	return s.Status == StatusActive || s.Status == StatusTrialing
}
