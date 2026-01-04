package handlers

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
)

type DashboardHandler struct {
	dashboardSvc *services.DashboardService
}

func NewDashboardHandler(dashboardSvc *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardSvc: dashboardSvc}
}

// GetDashboard returns full dashboard data
// @Summary Get dashboard data
// @Description Get aggregated dashboard data including summary stats, charts, recent activity
// @Tags Dashboard
// @Accept json
// @Produce json
// @Param workspace_id path string true "Workspace ID"
// @Param period query string false "Period for charts: 7d, 30d, 90d" default(7d)
// @Success 200 {object} services.DashboardSummary
// @Router /workspaces/{workspace_id}/dashboard [get]
func (h *DashboardHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "7d"
	}

	dashboard, err := h.dashboardSvc.GetDashboard(r.Context(), wsCtx.WorkspaceID, period)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to get dashboard data")
		return
	}

	wsID := wsCtx.WorkspaceID.String()
	dto.NewResponse(dashboard).
		WithLinks(&dto.Links{Self: "/api/v1/workspaces/" + wsID + "/dashboard"}).
		Send(w)
}

// GetQuickStats returns lightweight stats for sidebar/header
// @Summary Get quick stats
// @Description Get lightweight stats for sidebar or header display
// @Tags Dashboard
// @Accept json
// @Produce json
// @Param workspace_id path string true "Workspace ID"
// @Success 200 {object} services.QuickStats
// @Router /workspaces/{workspace_id}/stats [get]
func (h *DashboardHandler) GetQuickStats(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	stats, err := h.dashboardSvc.GetQuickStats(r.Context(), wsCtx.WorkspaceID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to get stats")
		return
	}

	wsID := wsCtx.WorkspaceID.String()
	dto.NewResponse(stats).
		WithLinks(&dto.Links{Self: "/api/v1/workspaces/" + wsID + "/stats"}).
		Send(w)
}
