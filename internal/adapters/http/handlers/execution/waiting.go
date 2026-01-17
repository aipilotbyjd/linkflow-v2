package execution

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type WaitingExecution struct {
	ID           string     `json:"id"`
	WorkflowID   string     `json:"workflowId"`
	WorkflowName string     `json:"workflowName"`
	NodeID       string     `json:"nodeId"`
	NodeName     string     `json:"nodeName"`
	WaitType     string     `json:"waitType"`
	WaitUntil    *time.Time `json:"waitUntil,omitempty"`
	ResumeToken  string     `json:"resumeToken"`
	CreatedAt    time.Time  `json:"createdAt"`
}

type GetWaitingHandler struct {
	executionRepo execution.Repository
}

func NewGetWaitingHandler(executionRepo execution.Repository) *GetWaitingHandler {
	return &GetWaitingHandler{executionRepo: executionRepo}
}

func (h *GetWaitingHandler) Handle(w http.ResponseWriter, r *http.Request) {
	executionIDStr := chi.URLParam(r, "executionId")
	executionID, err := uuid.Parse(executionIDStr)
	if err != nil {
		common.BadRequest(w, "invalid execution ID")
		return
	}

	workspaceID := middleware.GetWorkspaceID(r.Context())

	exec, err := h.executionRepo.FindByID(r.Context(), executionID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	if exec.WorkspaceID.String() != workspaceID.String() {
		common.Forbidden(w, "access denied")
		return
	}

	if exec.Status != execution.StatusWaiting {
		common.Error(w, http.StatusBadRequest, "NOT_WAITING", "Execution is not in waiting state")
		return
	}

	waiting := WaitingExecution{
		ID:           exec.ID.String(),
		WorkflowID:   exec.WorkflowID.String(),
		WorkflowName: "Workflow",
		NodeID:       "wait-node",
		NodeName:     "Wait Node",
		WaitType:     "manual",
		ResumeToken:  uuid.New().String(),
		CreatedAt:    exec.CreatedAt,
	}

	common.Success(w, waiting)
}

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
