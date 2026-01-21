package dashboard

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/credential"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
	"github.com/linkflow-ai/linkflow/internal/core/domain/schedule"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
)

type QuickStatsHandler struct {
	workflowRepo   workflow.Repository
	executionRepo  execution.Repository
	credentialRepo credential.Repository
	scheduleRepo   schedule.Repository
}

func NewQuickStatsHandler(
	workflowRepo workflow.Repository,
	executionRepo execution.Repository,
	credentialRepo credential.Repository,
	scheduleRepo schedule.Repository,
) *QuickStatsHandler {
	return &QuickStatsHandler{
		workflowRepo:   workflowRepo,
		executionRepo:  executionRepo,
		credentialRepo: credentialRepo,
		scheduleRepo:   scheduleRepo,
	}
}

type QuickStatsResponse struct {
	Workflows   WorkflowQuickStats   `json:"workflows"`
	Executions  ExecutionQuickStats  `json:"executions"`
	Credentials CredentialQuickStats `json:"credentials"`
	Schedules   ScheduleQuickStats   `json:"schedules"`
}

type WorkflowQuickStats struct {
	Total  int64 `json:"total"`
	Active int64 `json:"active"`
}

type ExecutionQuickStats struct {
	Running int64 `json:"running"`
	Queued  int64 `json:"queued"`
	Today   int64 `json:"today"`
}

type CredentialQuickStats struct {
	Total        int64 `json:"total"`
	ExpiringSoon int64 `json:"expiring_soon"`
}

type ScheduleQuickStats struct {
	Total  int64 `json:"total"`
	Active int64 `json:"active"`
}

func (h *QuickStatsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		common.BadRequest(w, "workspace context required")
		return
	}

	response := QuickStatsResponse{}

	// Get workflow stats
	workflows, total, err := h.workflowRepo.FindByWorkspaceID(r.Context(), wsCtx.WorkspaceID, nil)
	if err != nil {
		common.HandleError(w, err)
		return
	}
	response.Workflows.Total = total
	for _, wf := range workflows {
		if wf.Status == workflow.StatusActive {
			response.Workflows.Active++
		}
	}

	// Get execution stats
	executions, _, err := h.executionRepo.FindByWorkspaceID(r.Context(), wsCtx.WorkspaceID, nil)
	if err != nil {
		common.HandleError(w, err)
		return
	}
	for _, exec := range executions {
		switch exec.Status {
		case execution.StatusRunning:
			response.Executions.Running++
		case execution.StatusQueued:
			response.Executions.Queued++
		}
	}

	// Get credential stats
	_, credTotal, err := h.credentialRepo.FindByWorkspaceID(r.Context(), wsCtx.WorkspaceID, nil)
	if err != nil {
		common.HandleError(w, err)
		return
	}
	response.Credentials.Total = credTotal

	// Get schedule stats
	schedules, schTotal, err := h.scheduleRepo.FindByWorkspaceID(r.Context(), wsCtx.WorkspaceID, nil)
	if err != nil {
		common.HandleError(w, err)
		return
	}
	response.Schedules.Total = schTotal
	for _, sch := range schedules {
		if sch.IsActive {
			response.Schedules.Active++
		}
	}

	common.Success(w, response)
}
