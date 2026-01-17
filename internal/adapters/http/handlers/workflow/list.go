package workflow

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	workflowQuery "github.com/linkflow-ai/linkflow/internal/core/application/query/workflow"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// ListHandler handles workflow listing
type ListHandler struct {
	handler *workflowQuery.ListWorkflowsHandler
}

// NewListHandler creates a new list handler
func NewListHandler(handler *workflowQuery.ListWorkflowsHandler) *ListHandler {
	return &ListHandler{handler: handler}
}

// Handle handles the list workflows request
func (h *ListHandler) Handle(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		common.BadRequest(w, "workspace context required")
		return
	}

	// Parse query parameters
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

	var status *workflow.Status
	if s := query.Get("status"); s != "" {
		parsed := workflow.Status(s)
		if parsed.IsValid() {
			status = &parsed
		}
	}

	var folderID *uuid.UUID
	if f := query.Get("folder_id"); f != "" {
		if parsed, err := uuid.Parse(f); err == nil {
			folderID = &parsed
		}
	}

	var isFavorite *bool
	if fav := query.Get("is_favorite"); fav != "" {
		parsed := fav == "true"
		isFavorite = &parsed
	}

	search := query.Get("search")
	tags := query["tags"]

	result, err := h.handler.Handle(r.Context(), workflowQuery.ListWorkflowsQuery{
		WorkspaceID: wsCtx.WorkspaceID,
		Status:      status,
		FolderID:    folderID,
		IsFavorite:  isFavorite,
		Search:      search,
		Tags:        tags,
		Page:        page,
		PageSize:    pageSize,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	// Convert to response
	workflows := make([]WorkflowResponse, len(result.Workflows))
	for i, wf := range result.Workflows {
		workflows[i] = ToWorkflowResponse(&wf)
	}

	common.List(w, workflows, types.PageResponse{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalItems: result.Total,
		TotalPages: result.TotalPages,
		HasMore:    result.Page < result.TotalPages,
	})
}
