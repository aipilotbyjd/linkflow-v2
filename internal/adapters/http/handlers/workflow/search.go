package workflow

import (
	"net/http"
	"strconv"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type SearchHandler struct {
	workflowRepo workflow.Repository
}

func NewSearchHandler(workflowRepo workflow.Repository) *SearchHandler {
	return &SearchHandler{workflowRepo: workflowRepo}
}

func (h *SearchHandler) Handle(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		common.BadRequest(w, "workspace context required")
		return
	}

	query := r.URL.Query().Get("q")
	status := r.URL.Query().Get("status")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	opts := &workflow.ListOptions{
		ListOptions: types.NewListOptions(page, pageSize),
		Search:      query,
	}
	if status != "" {
		s := workflow.Status(status)
		opts.Status = &s
	}

	workflows, total, err := h.workflowRepo.FindByWorkspaceID(r.Context(), wsCtx.WorkspaceID, opts)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	var responses []WorkflowResponse
	for _, wf := range workflows {
		responses = append(responses, ToWorkflowResponse(&wf))
	}

	common.List(w, responses, types.PageResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: total,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
		HasMore:    int64(page*pageSize) < total,
	})
}
