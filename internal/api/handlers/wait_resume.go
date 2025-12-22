package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
)

type WaitResumeHandler struct {
	waitResumeMgr *services.WaitResumeManager
	waitingRepo   *repositories.WaitingExecutionRepository
}

func NewWaitResumeHandler(
	waitResumeMgr *services.WaitResumeManager,
	waitingRepo *repositories.WaitingExecutionRepository,
) *WaitResumeHandler {
	return &WaitResumeHandler{
		waitResumeMgr: waitResumeMgr,
		waitingRepo:   waitingRepo,
	}
}

// ListWaiting returns waiting executions for a workspace
func (h *WaitResumeHandler) ListWaiting(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		dto.ErrorResponse(w, http.StatusUnauthorized, "workspace required")
		return
	}

	opts := &repositories.ListOptions{Offset: 0, Limit: 50}
	waitings, total, err := h.waitingRepo.FindPendingByWorkspace(r.Context(), wsCtx.WorkspaceID, opts)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"waiting_executions": waitings,
		"total":              total,
	})
}

// GetByExecution returns waiting executions for a specific execution
func (h *WaitResumeHandler) GetByExecution(w http.ResponseWriter, r *http.Request) {
	executionIDStr := chi.URLParam(r, "executionID")
	executionID, err := uuid.Parse(executionIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid execution ID")
		return
	}

	waitings, err := h.waitResumeMgr.GetWaitingByExecution(r.Context(), executionID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"waiting_executions": waitings,
		"count":              len(waitings),
	})
}

// ResumeRequest represents a request to resume an execution
type ResumeRequest struct {
	Data models.JSON `json:"data,omitempty"`
}

// Resume resumes a waiting execution
func (h *WaitResumeHandler) Resume(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		dto.ErrorResponse(w, http.StatusBadRequest, "resume token required")
		return
	}

	var req ResumeRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	waiting, err := h.waitResumeMgr.ResumeExecution(r.Context(), token, req.Data)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"message":      "Execution resumed",
		"execution_id": waiting.ExecutionID,
		"workflow_id":  waiting.WorkflowID,
		"node_id":      waiting.NodeID,
	})
}

// GetWaitingStatus returns status of a waiting execution by token
func (h *WaitResumeHandler) GetWaitingStatus(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		dto.ErrorResponse(w, http.StatusBadRequest, "resume token required")
		return
	}

	waiting, err := h.waitResumeMgr.GetWaitingExecution(r.Context(), token)
	if err != nil {
		dto.ErrorResponse(w, http.StatusNotFound, "waiting execution not found")
		return
	}

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"execution_id": waiting.ExecutionID,
		"workflow_id":  waiting.WorkflowID,
		"node_id":      waiting.NodeID,
		"status":       waiting.Status,
		"timeout_at":   waiting.TimeoutAt,
		"created_at":   waiting.CreatedAt,
	})
}
