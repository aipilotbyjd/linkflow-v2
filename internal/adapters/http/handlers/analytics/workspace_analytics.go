package analytics

import (
	"net/http"
	"time"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	analyticsQry "github.com/linkflow-ai/linkflow/internal/core/application/query/analytics"
)

type WorkspaceAnalyticsHandler struct {
	handler *analyticsQry.GetWorkspaceAnalyticsHandler
}

func NewWorkspaceAnalyticsHandler(handler *analyticsQry.GetWorkspaceAnalyticsHandler) *WorkspaceAnalyticsHandler {
	return &WorkspaceAnalyticsHandler{handler: handler}
}

func (h *WorkspaceAnalyticsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workspaceID := middleware.GetWorkspaceID(r.Context())

	query := r.URL.Query()

	startDate := time.Now().AddDate(0, 0, -30)
	if s := query.Get("start_date"); s != "" {
		if parsed, err := time.Parse(time.DateOnly, s); err == nil {
			startDate = parsed
		}
	}

	endDate := time.Now()
	if e := query.Get("end_date"); e != "" {
		if parsed, err := time.Parse(time.DateOnly, e); err == nil {
			endDate = parsed
		}
	}

	result, err := h.handler.Handle(r.Context(), analyticsQry.GetWorkspaceAnalyticsQuery{
		WorkspaceID: workspaceID,
		StartDate:   startDate,
		EndDate:     endDate,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, ToWorkspaceAnalyticsResponse(result))
}
