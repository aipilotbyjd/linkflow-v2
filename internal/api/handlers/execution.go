package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/queue"
)

type ExecutionHandler struct {
	executionSvc *services.ExecutionService
	queueClient  *queue.Client
}

func NewExecutionHandler(executionSvc *services.ExecutionService, queueClient *queue.Client) *ExecutionHandler {
	return &ExecutionHandler{
		executionSvc: executionSvc,
		queueClient:  queueClient,
	}
}

func (h *ExecutionHandler) List(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "workspace context required")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	opts := repositories.NewListOptions(page, perPage)

	executions, total, err := h.executionSvc.GetByWorkspace(r.Context(), wsCtx.WorkspaceID, opts)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to list executions")
		return
	}

	var response []dto.ExecutionResponse
	for _, exec := range executions {
		var startedAt, completedAt *int64
		if exec.StartedAt != nil {
			ts := exec.StartedAt.Unix()
			startedAt = &ts
		}
		if exec.CompletedAt != nil {
			ts := exec.CompletedAt.Unix()
			completedAt = &ts
		}

		response = append(response, dto.ExecutionResponse{
			ID:              exec.ID.String(),
			WorkflowID:      exec.WorkflowID.String(),
			WorkflowVersion: exec.WorkflowVersion,
			Status:          exec.Status,
			TriggerType:     exec.TriggerType,
			InputData:       exec.InputData,
			OutputData:      exec.OutputData,
			ErrorMessage:    exec.ErrorMessage,
			ErrorNodeID:     exec.ErrorNodeID,
			NodesTotal:      exec.NodesTotal,
			NodesCompleted:  exec.NodesCompleted,
			QueuedAt:        exec.QueuedAt.Unix(),
			StartedAt:       startedAt,
			CompletedAt:     completedAt,
		})
	}

	totalPages := int(total) / opts.Limit
	if int(total)%opts.Limit > 0 {
		totalPages++
	}

	dto.JSONWithMeta(w, http.StatusOK, response, &dto.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	})
}

func (h *ExecutionHandler) Get(w http.ResponseWriter, r *http.Request) {
	executionIDStr := chi.URLParam(r, "executionID")
	executionID, err := uuid.Parse(executionIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid execution ID")
		return
	}

	execution, err := h.executionSvc.GetByID(r.Context(), executionID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusNotFound, "execution not found")
		return
	}

	var startedAt, completedAt *int64
	if execution.StartedAt != nil {
		ts := execution.StartedAt.Unix()
		startedAt = &ts
	}
	if execution.CompletedAt != nil {
		ts := execution.CompletedAt.Unix()
		completedAt = &ts
	}

	dto.JSON(w, http.StatusOK, dto.ExecutionResponse{
		ID:              execution.ID.String(),
		WorkflowID:      execution.WorkflowID.String(),
		WorkflowVersion: execution.WorkflowVersion,
		Status:          execution.Status,
		TriggerType:     execution.TriggerType,
		InputData:       execution.InputData,
		OutputData:      execution.OutputData,
		ErrorMessage:    execution.ErrorMessage,
		ErrorNodeID:     execution.ErrorNodeID,
		NodesTotal:      execution.NodesTotal,
		NodesCompleted:  execution.NodesCompleted,
		QueuedAt:        execution.QueuedAt.Unix(),
		StartedAt:       startedAt,
		CompletedAt:     completedAt,
	})
}

func (h *ExecutionHandler) GetNodes(w http.ResponseWriter, r *http.Request) {
	executionIDStr := chi.URLParam(r, "executionID")
	executionID, err := uuid.Parse(executionIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid execution ID")
		return
	}

	nodeExecutions, err := h.executionSvc.GetNodeExecutions(r.Context(), executionID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to get node executions")
		return
	}

	var response []dto.NodeExecutionResponse
	for _, ne := range nodeExecutions {
		var startedAt, completedAt *int64
		if ne.StartedAt != nil {
			ts := ne.StartedAt.Unix()
			startedAt = &ts
		}
		if ne.CompletedAt != nil {
			ts := ne.CompletedAt.Unix()
			completedAt = &ts
		}

		response = append(response, dto.NodeExecutionResponse{
			ID:           ne.ID.String(),
			NodeID:       ne.NodeID,
			NodeType:     ne.NodeType,
			NodeName:     ne.NodeName,
			Status:       ne.Status,
			InputData:    ne.InputData,
			OutputData:   ne.OutputData,
			ErrorMessage: ne.ErrorMessage,
			DurationMs:   ne.DurationMs,
			StartedAt:    startedAt,
			CompletedAt:  completedAt,
		})
	}

	dto.JSON(w, http.StatusOK, response)
}

func (h *ExecutionHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	executionIDStr := chi.URLParam(r, "executionID")
	executionID, err := uuid.Parse(executionIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid execution ID")
		return
	}

	if err := h.executionSvc.Cancel(r.Context(), executionID); err != nil {
		if err == services.ErrExecutionNotRunning {
			dto.ErrorResponse(w, http.StatusBadRequest, "execution is not running")
			return
		}
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to cancel execution")
		return
	}

	dto.JSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

func (h *ExecutionHandler) Retry(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		dto.ErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	executionIDStr := chi.URLParam(r, "executionID")
	executionID, err := uuid.Parse(executionIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid execution ID")
		return
	}

	execution, err := h.executionSvc.Retry(r.Context(), executionID, &claims.UserID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to retry execution")
		return
	}

	dto.Accepted(w, map[string]string{
		"execution_id": execution.ID.String(),
		"status":       "queued",
	})
}
