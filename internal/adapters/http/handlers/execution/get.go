package execution

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	executionQuery "github.com/linkflow-ai/linkflow/internal/core/application/query/execution"
)

// GetHandler handles getting a single execution
type GetHandler struct {
	handler *executionQuery.GetExecutionHandler
}

// NewGetHandler creates a new handler
func NewGetHandler(handler *executionQuery.GetExecutionHandler) *GetHandler {
	return &GetHandler{handler: handler}
}

// Handle handles the get execution request
func (h *GetHandler) Handle(w http.ResponseWriter, r *http.Request) {
	executionIDStr := chi.URLParam(r, "executionId")
	executionID, err := uuid.Parse(executionIDStr)
	if err != nil {
		common.BadRequest(w, "invalid execution ID")
		return
	}

	exec, err := h.handler.Handle(r.Context(), executionQuery.GetExecutionQuery{
		ExecutionID: executionID,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, ToExecutionResponse(exec))
}
