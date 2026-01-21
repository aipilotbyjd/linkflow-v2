package billing

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	billingapp "github.com/linkflow-ai/linkflow/internal/core/application/billing"
)

// DashboardHandler returns comprehensive usage dashboard data
type DashboardHandler struct {
	usageService *billingapp.UsageService
}

func NewDashboardHandler(usageService *billingapp.UsageService) *DashboardHandler {
	return &DashboardHandler{usageService: usageService}
}

func (h *DashboardHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceID := middleware.GetWorkspaceID(ctx)

	dashboard, err := h.usageService.GetUsageDashboard(ctx, workspaceID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, DashboardResponse{
		Plan: PlanInfo{
			ID:   dashboard.PlanID,
			Name: dashboard.PlanName,
		},
		Operations: UsageMetrics{
			Used:       dashboard.OperationsUsed,
			Limit:      dashboard.OperationsLimit,
			Percent:    dashboard.OperationsPercent,
			Projected:  dashboard.OperationsProjected,
			Rollover:   dashboard.OperationsRollover,
			Remaining:  dashboard.OperationsLimit - dashboard.OperationsUsed,
			IsUnlimited: dashboard.OperationsLimit < 0,
		},
		AICredits: UsageMetrics{
			Used:       dashboard.AICreditsUsed,
			Limit:      dashboard.AICreditsLimit,
			Percent:    dashboard.AICreditsPercent,
			Projected:  dashboard.AICreditsProjected,
			Rollover:   dashboard.AICreditsRollover,
			Remaining:  dashboard.AICreditsLimit - dashboard.AICreditsUsed,
			IsUnlimited: dashboard.AICreditsLimit < 0,
		},
		Overage: OverageInfo{
			Operations:   dashboard.OverageOperations,
			AICredits:    dashboard.OverageAICredits,
			ChargeCents:  dashboard.OverageChargeCents,
			ChargeFormatted: formatCents(dashboard.OverageChargeCents),
		},
		Period: PeriodInfo{
			Start:         dashboard.PeriodStart.Format("2006-01-02"),
			End:           dashboard.PeriodEnd.Format("2006-01-02"),
			DaysRemaining: dashboard.DaysRemaining,
		},
	})
}

// Response types

type DashboardResponse struct {
	Plan       PlanInfo     `json:"plan"`
	Operations UsageMetrics `json:"operations"`
	AICredits  UsageMetrics `json:"ai_credits"`
	Overage    OverageInfo  `json:"overage"`
	Period     PeriodInfo   `json:"period"`
}

type PlanInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type UsageMetrics struct {
	Used        int64   `json:"used"`
	Limit       int64   `json:"limit"`
	Percent     float64 `json:"percent"`
	Projected   int64   `json:"projected"`
	Rollover    int64   `json:"rollover"`
	Remaining   int64   `json:"remaining"`
	IsUnlimited bool    `json:"is_unlimited"`
}

type OverageInfo struct {
	Operations      int64  `json:"operations"`
	AICredits       int64  `json:"ai_credits"`
	ChargeCents     int64  `json:"charge_cents"`
	ChargeFormatted string `json:"charge_formatted"`
}

type PeriodInfo struct {
	Start         string `json:"start"`
	End           string `json:"end"`
	DaysRemaining int    `json:"days_remaining"`
}

func formatCents(cents int64) string {
	dollars := float64(cents) / 100
	if dollars == 0 {
		return "$0.00"
	}
	return "$" + formatFloat(dollars)
}

func formatFloat(f float64) string {
	if f == float64(int64(f)) {
		return string(rune(int64(f))) + ".00"
	}
	// Simple formatting - in production use strconv or fmt
	return string(rune(int64(f*100)/100)) + "." + string(rune(int64(f*100)%100))
}
