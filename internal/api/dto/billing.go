package dto

// Billing requests

type CreateSubscriptionRequest struct {
	PlanID       string `json:"plan_id" validate:"required"`
	BillingCycle string `json:"billing_cycle" validate:"required,oneof=monthly yearly"`
	PaymentToken string `json:"payment_token,omitempty"`
}

// Billing responses

type PlanResponse struct {
	ID               string      `json:"id"`
	Name             string      `json:"name"`
	Description      *string     `json:"description,omitempty"`
	PriceMonthly     int         `json:"price_monthly"`
	PriceYearly      int         `json:"price_yearly"`
	ExecutionsLimit  int         `json:"executions_limit"`
	WorkflowsLimit   int         `json:"workflows_limit"`
	MembersLimit     int         `json:"members_limit"`
	CredentialsLimit int         `json:"credentials_limit"`
	Features         interface{} `json:"features,omitempty"`
}

type SubscriptionResponse struct {
	ID                 string `json:"id"`
	PlanID             string `json:"plan_id"`
	Status             string `json:"status"`
	BillingCycle       string `json:"billing_cycle"`
	CurrentPeriodStart int64  `json:"current_period_start"`
	CurrentPeriodEnd   int64  `json:"current_period_end"`
	CancelAt           *int64 `json:"cancel_at,omitempty"`
}

type UsageResponse struct {
	// Core counts
	Executions   int   `json:"executions"`
	Workflows    int   `json:"workflows"`
	Members      int   `json:"members"`
	Credentials  int   `json:"credentials"`
	StorageBytes int64 `json:"storage_bytes"`

	// Credits
	CreditsUsed      int `json:"credits_used"`
	CreditsIncluded  int `json:"credits_included"`
	CreditsPurchased int `json:"credits_purchased"`
	CreditsRemaining int `json:"credits_remaining"`

	// Execution details
	ExecutionsSuccess int `json:"executions_success"`
	ExecutionsFailed  int `json:"executions_failed"`
	Operations        int `json:"operations"`

	// Webhooks & Schedules
	WebhooksCalled     int `json:"webhooks_called"`
	SchedulesTriggered int `json:"schedules_triggered"`
	Schedules          int `json:"schedules"`
	Webhooks           int `json:"webhooks"`

	// Data transfer
	DataTransferIn  int64 `json:"data_transfer_in"`
	DataTransferOut int64 `json:"data_transfer_out"`

	// Overage
	OverageCredits int `json:"overage_credits"`
	OverageCost    int `json:"overage_cost"`

	// Period
	PeriodStart int64 `json:"period_start"`
	PeriodEnd   int64 `json:"period_end"`
}
