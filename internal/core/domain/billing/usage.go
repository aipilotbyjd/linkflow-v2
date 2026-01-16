package billing

import (
	"time"

	"github.com/google/uuid"
)

type Usage struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID   uuid.UUID `gorm:"type:uuid;index;not null" json:"workspace_id"`
	PeriodStart   time.Time `gorm:"index;not null" json:"period_start"`
	PeriodEnd     time.Time `gorm:"index;not null" json:"period_end"`
	Executions    int64     `gorm:"default:0" json:"executions"`
	ApiCalls      int64     `gorm:"default:0" json:"api_calls"`
	StorageBytes  int64     `gorm:"default:0" json:"storage_bytes"`
	BandwidthBytes int64    `gorm:"default:0" json:"bandwidth_bytes"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
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

func (u *Usage) IncrementExecutions(count int64) {
	u.Executions += count
	u.UpdatedAt = time.Now()
}

func (u *Usage) IncrementApiCalls(count int64) {
	u.ApiCalls += count
	u.UpdatedAt = time.Now()
}
