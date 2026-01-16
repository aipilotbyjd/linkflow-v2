package billing

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
)

type UsageResponse struct {
	WorkspaceID     string `json:"workspace_id"`
	Period          string `json:"period"`
	ExecutionCount  int64  `json:"execution_count"`
	ExecutionLimit  int64  `json:"execution_limit"`
	WorkflowCount   int64  `json:"workflow_count"`
	WorkflowLimit   int64  `json:"workflow_limit"`
	CredentialCount int64  `json:"credential_count"`
	CredentialLimit int64  `json:"credential_limit"`
	StorageUsedMB   int64  `json:"storage_used_mb"`
	StorageLimitMB  int64  `json:"storage_limit_mb"`
}

type GetUsageHandler struct {
	usageRepo billing.UsageRepository
}

func NewGetUsageHandler(usageRepo billing.UsageRepository) *GetUsageHandler {
	return &GetUsageHandler{usageRepo: usageRepo}
}

func (h *GetUsageHandler) Handle(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		common.BadRequest(w, "workspace context required")
		return
	}

	usage, err := h.usageRepo.GetCurrentPeriodUsage(r.Context(), wsCtx.WorkspaceID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, UsageResponse{
		WorkspaceID:     wsCtx.WorkspaceID.String(),
		Period:          usage.PeriodStart.Format("2006-01") + " to " + usage.PeriodEnd.Format("2006-01"),
		ExecutionCount:  usage.Executions,
		ExecutionLimit:  0, // Determined by plan
		WorkflowCount:   0, // Need separate query
		WorkflowLimit:   0, // Determined by plan
		CredentialCount: 0, // Need separate query
		CredentialLimit: 0, // Determined by plan
		StorageUsedMB:   usage.StorageBytes / (1024 * 1024),
		StorageLimitMB:  0, // Determined by plan
	})
}
