package analytics

import (
	"net/http"
	"time"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	analyticsQuery "github.com/linkflow-ai/linkflow/internal/core/application/query/analytics"
)

type WorkspaceAnalyticsHandler struct {
	handler *analyticsQuery.GetWorkspaceAnalyticsHandler
}

func NewWorkspaceAnalyticsHandler(handler *analyticsQuery.GetWorkspaceAnalyticsHandler) *WorkspaceAnalyticsHandler {
	return &WorkspaceAnalyticsHandler{handler: handler}
}

func (h *WorkspaceAnalyticsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		common.BadRequest(w, "workspace context required")
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

	result, err := h.handler.Handle(r.Context(), analyticsQuery.GetWorkspaceAnalyticsQuery{
		WorkspaceID: wsCtx.WorkspaceID,
		StartDate:   startDate,
		EndDate:     endDate,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, result)
}
