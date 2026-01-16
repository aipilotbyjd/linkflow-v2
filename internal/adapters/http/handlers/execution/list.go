package execution

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	executionQuery "github.com/linkflow-ai/linkflow/internal/core/application/query/execution"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// ListHandler handles listing executions
type ListHandler struct {
	handler *executionQuery.ListExecutionsHandler
}

// NewListHandler creates a new handler
func NewListHandler(handler *executionQuery.ListExecutionsHandler) *ListHandler {
	return &ListHandler{handler: handler}
}

// Handle handles the list executions request
func (h *ListHandler) Handle(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		common.BadRequest(w, "workspace context required")
		return
	}

	query := r.URL.Query()

	page := 1
	if p := query.Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	pageSize := types.DefaultPageSize
	if ps := query.Get("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}

	var status *execution.Status
	if s := query.Get("status"); s != "" {
		parsed := execution.Status(s)
		if parsed.IsValid() {
			status = &parsed
		}
	}

	var workflowID *uuid.UUID
	if wf := query.Get("workflow_id"); wf != "" {
		if parsed, err := uuid.Parse(wf); err == nil {
			workflowID = &parsed
		}
	}

	var triggerType *string
	if tt := query.Get("trigger_type"); tt != "" {
		triggerType = &tt
	}

	result, err := h.handler.Handle(r.Context(), executionQuery.ListExecutionsQuery{
		WorkspaceID: wsCtx.WorkspaceID,
		WorkflowID:  workflowID,
		Status:      status,
		TriggerType: triggerType,
		Page:        page,
		PageSize:    pageSize,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	executions := make([]ExecutionResponse, len(result.Executions))
	for i, e := range result.Executions {
		executions[i] = toExecutionResponse(&e)
	}

	common.List(w, executions, types.PageResponse{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalItems: result.Total,
		TotalPages: result.TotalPages,
		HasMore:    result.Page < result.TotalPages,
	})
}
