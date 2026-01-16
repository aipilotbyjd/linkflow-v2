package schedule

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	scheduleQuery "github.com/linkflow-ai/linkflow/internal/core/application/query/schedule"
)

type GetHandler struct {
	handler *scheduleQuery.GetScheduleHandler
}

func NewGetHandler(handler *scheduleQuery.GetScheduleHandler) *GetHandler {
	return &GetHandler{handler: handler}
}

func (h *GetHandler) Handle(w http.ResponseWriter, r *http.Request) {
	scheduleIDStr := chi.URLParam(r, "scheduleId")
	scheduleID, err := uuid.Parse(scheduleIDStr)
	if err != nil {
		common.BadRequest(w, "invalid schedule ID")
		return
	}

	sched, err := h.handler.Handle(r.Context(), scheduleQuery.GetScheduleQuery{
		ScheduleID: scheduleID,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, toScheduleResponse(sched))
}
