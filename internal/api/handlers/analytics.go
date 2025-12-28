package handlers

import (
	"net/http"
	"time"

	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
)

type AnalyticsHandler struct {
	analyticsSvc *services.AnalyticsService
}

func NewAnalyticsHandler(analyticsSvc *services.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsSvc: analyticsSvc}
}

func (h *AnalyticsHandler) GetWorkspaceAnalytics(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	start, end := parseDateRange(r)
	analytics, err := h.analyticsSvc.GetWorkspaceAnalytics(r.Context(), wsCtx.WorkspaceID, start, end)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to get analytics")
		return
	}

	dto.OK(w, analytics)
}

func (h *AnalyticsHandler) GetWorkflowAnalytics(w http.ResponseWriter, r *http.Request) {
	workflowID, ok := middleware.ParseUUID(w, r, "workflowId")
	if !ok {
		return
	}

	start, end := parseDateRange(r)
	analytics, err := h.analyticsSvc.GetWorkflowAnalytics(r.Context(), workflowID, start, end)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to get analytics")
		return
	}

	dto.OK(w, analytics)
}

func parseDateRange(r *http.Request) (time.Time, time.Time) {
	end := time.Now()
	start := end.AddDate(0, 0, -30)

	if s := dto.QueryString(r, "start"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			start = t
		}
	}
	if e := dto.QueryString(r, "end"); e != "" {
		if t, err := time.Parse("2006-01-02", e); err == nil {
			end = t
		}
	}
	return start, end
}
