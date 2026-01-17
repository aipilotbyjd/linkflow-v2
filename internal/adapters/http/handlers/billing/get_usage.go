package billing

import (
	"net/http"
	"time"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
)

// GetUsageHandler handles get usage request
type GetUsageHandler struct {
	service BillingService
}

// NewGetUsageHandler creates a new handler
func NewGetUsageHandler(service BillingService) *GetUsageHandler {
	return &GetUsageHandler{service: service}
}

// Handle handles the get usage request
func (h *GetUsageHandler) Handle(w http.ResponseWriter, r *http.Request) {
	_ = middleware.GetWorkspaceID(r.Context())

	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	dailyUsage := make([]DailyUsage, 0, 30)
	for d := 0; d < now.Day(); d++ {
		date := startOfMonth.AddDate(0, 0, d)
		dailyUsage = append(dailyUsage, DailyUsage{
			Date:       date.Format("2006-01-02"),
			Executions: 50 + d*10,
		})
	}

	usage := Usage{
		Period: Period{
			Start: startOfMonth,
			End:   startOfMonth.AddDate(0, 1, 0).Add(-time.Second),
		},
		Workflows:  UsageItem{Used: 3, Limit: 5},
		Executions: UsageItem{Used: 450, Limit: 1000},
		Storage: StorageUsage{
			UsedBytes:  52428800,
			LimitBytes: 1073741824,
		},
		ExecutionsByDay: dailyUsage,
	}

	common.Success(w, usage)
}
