package analytics

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	analyticsQuery "github.com/linkflow-ai/linkflow/internal/core/application/query/analytics"
)

type WorkflowAnalyticsHandler struct {
	handler *analyticsQuery.GetWorkflowAnalyticsHandler
}

func NewWorkflowAnalyticsHandler(handler *analyticsQuery.GetWorkflowAnalyticsHandler) *WorkflowAnalyticsHandler {
	return &WorkflowAnalyticsHandler{handler: handler}
}

func (h *WorkflowAnalyticsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workflowIDStr := chi.URLParam(r, "workflowId")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		common.BadRequest(w, "invalid workflow ID")
		return
	}

	query := r.URL.Query()

	startDate := time.Now().AddDate(0, 0, -30)
	if s := query.Get("start_date"); s != "" {
		if parsed, err := time.Parse("2006-01-02", s); err == nil {
			startDate = parsed
		}
	}

	endDate := time.Now()
	if e := query.Get("end_date"); e != "" {
		if parsed, err := time.Parse("2006-01-02", e); err == nil {
			endDate = parsed
		}
	}

	result, err := h.handler.Handle(r.Context(), analyticsQuery.GetWorkflowAnalyticsQuery{
		WorkflowID: workflowID,
		StartDate:  startDate,
		EndDate:    endDate,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, result)
}
