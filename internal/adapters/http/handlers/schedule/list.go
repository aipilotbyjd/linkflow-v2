package schedule

import (
	"net/http"
	"strconv"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	scheduleQuery "github.com/linkflow-ai/linkflow/internal/core/application/query/schedule"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type ListHandler struct {
	handler *scheduleQuery.ListSchedulesHandler
}

func NewListHandler(handler *scheduleQuery.ListSchedulesHandler) *ListHandler {
	return &ListHandler{handler: handler}
}

func (h *ListHandler) Handle(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		common.BadRequest(w, "workspace context required")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	result, err := h.handler.Handle(r.Context(), scheduleQuery.ListSchedulesQuery{
		WorkspaceID: wsCtx.WorkspaceID,
		Page:        page,
		PageSize:    pageSize,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	responses := make([]ScheduleResponse, 0, len(result.Schedules))
	for _, s := range result.Schedules {
		responses = append(responses, ToScheduleResponse(&s))
	}

	common.List(w, responses, types.PageResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: result.Total,
		TotalPages: result.TotalPages,
		HasMore:    int64(page*pageSize) < result.Total,
	})
}
