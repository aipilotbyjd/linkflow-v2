package execution

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	executionCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/execution"
)

type RetryHandler struct {
	handler *executionCmd.StartExecutionHandler
}

func NewRetryHandler(handler *executionCmd.StartExecutionHandler) *RetryHandler {
	return &RetryHandler{handler: handler}
}

func (h *RetryHandler) Handle(w http.ResponseWriter, r *http.Request) {
	executionIDStr := chi.URLParam(r, "executionId")
	_, err := uuid.Parse(executionIDStr)
	if err != nil {
		common.BadRequest(w, "invalid execution ID")
		return
	}

	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		common.BadRequest(w, "workspace context required")
		return
	}

	userClaims := middleware.GetUserFromContext(r.Context())
	if userClaims == nil {
		common.Unauthorized(w, "")
		return
	}

	// TODO: Get original execution, create new execution with same input
	common.Success(w, map[string]string{"status": "retry queued"})
}
