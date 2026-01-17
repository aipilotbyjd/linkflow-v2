package execution

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type ListWaitingHandler struct {
	executionRepo execution.Repository
}

func NewListWaitingHandler(executionRepo execution.Repository) *ListWaitingHandler {
	return &ListWaitingHandler{executionRepo: executionRepo}
}

func (h *ListWaitingHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workspaceID := middleware.GetWorkspaceID(r.Context())

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	opts := execution.NewListOptions(page, pageSize)
	opts.Status = func() *execution.Status { s := execution.StatusWaiting; return &s }()

	executions, total, err := h.executionRepo.FindByWorkspaceID(r.Context(), workspaceID, opts)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	var waiting []WaitingExecution
	for _, exec := range executions {
		waiting = append(waiting, WaitingExecution{
			ID:           exec.ID.String(),
			WorkflowID:   exec.WorkflowID.String(),
			WorkflowName: "Workflow",
			NodeID:       "wait-node",
			NodeName:     "Wait Node",
			WaitType:     "manual",
			ResumeToken:  uuid.New().String(),
			CreatedAt:    exec.CreatedAt,
		})
	}

	common.List(w, waiting, types.PageResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: total,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
		HasMore:    int64(page*pageSize) < total,
	})
}
