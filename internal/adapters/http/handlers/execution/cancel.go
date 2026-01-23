package execution

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
)

type CancelHandler struct {
	executionRepo execution.Repository
}

func NewCancelHandler(executionRepo execution.Repository) *CancelHandler {
	return &CancelHandler{executionRepo: executionRepo}
}

func (h *CancelHandler) Handle(w http.ResponseWriter, r *http.Request) {
	executionIDStr := chi.URLParam(r, "executionId")
	executionID, err := uuid.Parse(executionIDStr)
	if err != nil {
		common.BadRequest(w, "invalid execution ID")
		return
	}

	exec, err := h.executionRepo.FindByID(r.Context(), executionID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	if err := exec.Cancel(); err != nil {
		common.HandleError(w, err)
		return
	}

	if err := h.executionRepo.Update(r.Context(), exec); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, ToExecutionResponse(exec))
}
