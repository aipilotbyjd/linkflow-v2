package billing

import (
	"time"

	"github.com/google/uuid"
)

// Usage tracks workspace resource consumption (Make.com style - operations based)
type Usage struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID    uuid.UUID `gorm:"type:uuid;index;not null" json:"workspace_id"`
	PeriodStart    time.Time `gorm:"index;not null" json:"period_start"`
	PeriodEnd      time.Time `gorm:"index;not null" json:"period_end"`
	
	// Operations (Make.com style - each node execution = 1 operation)
	Operations     int64     `gorm:"default:0" json:"operations"`
	
	// AI Credits (separate pool for AI features)
	AICreditsUsed  int64     `gorm:"default:0" json:"ai_credits_used"`
	
	// Legacy fields (kept for compatibility)
	Executions     int64     `gorm:"default:0" json:"executions"`      // workflow runs
	ApiCalls       int64     `gorm:"default:0" json:"api_calls"`
	
	// Data metrics
	DataTransferMB int64     `gorm:"default:0" json:"data_transfer_mb"`
	StorageBytes   int64     `gorm:"default:0" json:"storage_bytes"`
	BandwidthBytes int64     `gorm:"default:0" json:"bandwidth_bytes"`
	
	// Active scenarios count
	ActiveScenarios int      `gorm:"default:0" json:"active_scenarios"`
	
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (Usage) TableName() string {
	return "usage_records"
}

func NewUsage(workspaceID uuid.UUID, periodStart, periodEnd time.Time) *Usage {
	now := time.Now()
	return &Usage{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// IncrementOperations adds to operations count (1 per node execution)
func (u *Usage) IncrementOperations(count int64) {
	u.Operations += count
	u.UpdatedAt = time.Now()
}

// IncrementAICredits adds to AI credits used
func (u *Usage) IncrementAICredits(count int64) {
	u.AICreditsUsed += count
	u.UpdatedAt = time.Now()
}

func (u *Usage) IncrementExecutions(count int64) {
	u.Executions += count
	u.UpdatedAt = time.Now()
}

func (u *Usage) IncrementApiCalls(count int64) {
	u.ApiCalls += count
	u.UpdatedAt = time.Now()
}

func (u *Usage) IncrementDataTransfer(mb int64) {
	u.DataTransferMB += mb
	u.UpdatedAt = time.Now()
}

// CheckLimits checks if usage exceeds plan limits
func (u *Usage) CheckLimits(limits *Limits) *UsageLimitStatus {
	status := &UsageLimitStatus{}
	
	if limits.OperationsPerMonth > 0 {
		status.OperationsUsed = u.Operations
		status.OperationsLimit = int64(limits.OperationsPerMonth)
		status.OperationsPercent = float64(u.Operations) / float64(limits.OperationsPerMonth) * 100
		status.OperationsExceeded = u.Operations > int64(limits.OperationsPerMonth)
	}
	
	if limits.AICreditsPerMonth > 0 {
		status.AICreditsUsed = u.AICreditsUsed
		status.AICreditsLimit = int64(limits.AICreditsPerMonth)
		status.AICreditsPercent = float64(u.AICreditsUsed) / float64(limits.AICreditsPerMonth) * 100
		status.AICreditsExceeded = u.AICreditsUsed > int64(limits.AICreditsPerMonth)
	}
	
	if limits.DataTransferMB > 0 {
		status.DataTransferUsed = u.DataTransferMB
		status.DataTransferLimit = int64(limits.DataTransferMB)
		status.DataTransferExceeded = u.DataTransferMB > int64(limits.DataTransferMB)
	}
	
	return status
}

// UsageLimitStatus represents current usage vs limits
type UsageLimitStatus struct {
	OperationsUsed     int64   `json:"operations_used"`
	OperationsLimit    int64   `json:"operations_limit"`
	OperationsPercent  float64 `json:"operations_percent"`
	OperationsExceeded bool    `json:"operations_exceeded"`
	
	AICreditsUsed      int64   `json:"ai_credits_used"`
	AICreditsLimit     int64   `json:"ai_credits_limit"`
	AICreditsPercent   float64 `json:"ai_credits_percent"`
	AICreditsExceeded  bool    `json:"ai_credits_exceeded"`
	
	DataTransferUsed     int64 `json:"data_transfer_used_mb"`
	DataTransferLimit    int64 `json:"data_transfer_limit_mb"`
	DataTransferExceeded bool  `json:"data_transfer_exceeded"`
}
