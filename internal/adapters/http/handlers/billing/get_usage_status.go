package billing

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	billingapp "github.com/linkflow-ai/linkflow/internal/core/application/billing"
	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
)

// GetUsageStatusHandler returns current usage status with limits
type GetUsageStatusHandler struct {
	usageService *billingapp.UsageService
}

func NewGetUsageStatusHandler(usageService *billingapp.UsageService) *GetUsageStatusHandler {
	return &GetUsageStatusHandler{usageService: usageService}
}

func (h *GetUsageStatusHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceID := middleware.GetWorkspaceID(ctx)

	status, err := h.usageService.GetUsageStatus(ctx, workspaceID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, UsageStatusResponse{
		Operations: UsageMetricResponse{
			Used:       status.OperationsUsed,
			Limit:      status.OperationsLimit,
			Percent:    status.OperationsPercent,
			Exceeded:   status.OperationsExceeded,
			Remaining:  status.OperationsLimit - status.OperationsUsed,
		},
		AICredits: UsageMetricResponse{
			Used:       status.AICreditsUsed,
			Limit:      status.AICreditsLimit,
			Percent:    status.AICreditsPercent,
			Exceeded:   status.AICreditsExceeded,
			Remaining:  status.AICreditsLimit - status.AICreditsUsed,
		},
		DataTransfer: DataTransferResponse{
			UsedMB:   status.DataTransferUsed,
			LimitMB:  status.DataTransferLimit,
			Exceeded: status.DataTransferExceeded,
		},
		OverageRates: OverageRatesResponse{
			OperationsPer1000: billing.DefaultOverageRates.OperationsPer1000,
			AICreditsPerUnit:  billing.DefaultOverageRates.AICreditsPerCredit,
		},
	})
}

// Response types

type UsageStatusResponse struct {
	Operations   UsageMetricResponse   `json:"operations"`
	AICredits    UsageMetricResponse   `json:"aiCredits"`
	DataTransfer DataTransferResponse  `json:"dataTransfer"`
	OverageRates OverageRatesResponse  `json:"overageRates"`
}

type UsageMetricResponse struct {
	Used      int64   `json:"used"`
	Limit     int64   `json:"limit"`
	Percent   float64 `json:"percent"`
	Exceeded  bool    `json:"exceeded"`
	Remaining int64   `json:"remaining"`
}

type DataTransferResponse struct {
	UsedMB   int64 `json:"usedMB"`
	LimitMB  int64 `json:"limitMB"`
	Exceeded bool  `json:"exceeded"`
}

type OverageRatesResponse struct {
	OperationsPer1000 int64 `json:"operationsPer1000Cents"`
	AICreditsPerUnit  int64 `json:"aiCreditsPerUnitCents"`
}
