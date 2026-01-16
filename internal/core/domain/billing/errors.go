package billing

import "errors"

var (
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrUsageNotFound        = errors.New("usage record not found")
	ErrInvoiceNotFound      = errors.New("invoice not found")
	ErrPlanNotFound         = errors.New("plan not found")
	ErrSubscriptionInactive = errors.New("subscription is not active")
	ErrLimitExceeded        = errors.New("plan limit exceeded")
	ErrPaymentFailed        = errors.New("payment failed")
)
