package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/pkg/queue"
	"github.com/linkflow-ai/linkflow/internal/pkg/validator"
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

	response := []dto.ExecutionResponse{}
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

	// SECURITY: Validate workspace ownership to prevent cross-tenant access
	if !ValidateWorkspaceOwnership(w, r, execution) {
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

	// SECURITY: Validate ownership before accessing node data
	existing, err := h.executionSvc.GetByID(r.Context(), executionID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusNotFound, "execution not found")
		return
	}
	if !ValidateWorkspaceOwnership(w, r, existing) {
		return
	}

	nodeExecutions, err := h.executionSvc.GetNodeExecutions(r.Context(), executionID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to get node executions")
		return
	}

	response := []dto.NodeExecutionResponse{}
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
		dto.BadRequest(w, "invalid execution ID")
		return
	}

	// SECURITY: Validate ownership before cancellation
	existing, err := h.executionSvc.GetByID(r.Context(), executionID)
	if err != nil {
		dto.NotFound(w, "Execution")
		return
	}
	if !ValidateWorkspaceOwnership(w, r, existing) {
		return
	}

	// Business rule: Check if execution can be cancelled
	if err := validator.CanCancelExecution(existing.Status); err != nil {
		dto.BadRequest(w, err.Error())
		return
	}

	if err := h.executionSvc.Cancel(r.Context(), executionID); err != nil {
		if err == services.ErrExecutionNotRunning {
			dto.BadRequest(w, "execution is not running")
			return
		}
		dto.InternalServerError(w, "failed to cancel execution")
		return
	}

	dto.OK(w, map[string]string{"status": "cancelled"})
}

func (h *ExecutionHandler) Retry(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		dto.Unauthorized(w, "unauthorized")
		return
	}

	executionIDStr := chi.URLParam(r, "executionID")
	executionID, err := uuid.Parse(executionIDStr)
	if err != nil {
		dto.BadRequest(w, "invalid execution ID")
		return
	}

	// SECURITY: Validate ownership before retry
	existing, err := h.executionSvc.GetByID(r.Context(), executionID)
	if err != nil {
		dto.NotFound(w, "Execution")
		return
	}
	if !ValidateWorkspaceOwnership(w, r, existing) {
		return
	}

	// Business rule: Check if execution can be retried
	if err := validator.CanRetryExecution(existing.Status); err != nil {
		dto.BadRequest(w, err.Error())
		return
	}

	execution, err := h.executionSvc.Retry(r.Context(), executionID, &claims.UserID)
	if err != nil {
		dto.InternalServerError(w, "failed to retry execution")
		return
	}

	dto.Accepted(w, map[string]string{
		"execution_id": execution.ID.String(),
		"status":       "queued",
	})
}

// Search searches executions with filters
func (h *ExecutionHandler) Search(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "workspace context required")
		return
	}

	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	opts := repositories.NewListOptions(page, perPage)

	// Build filter
	filter := repositories.ExecutionFilter{
		WorkspaceID: &wsCtx.WorkspaceID,
	}

	// Status filter
	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = &status
	}

	// Workflow filter
	if workflowIDStr := r.URL.Query().Get("workflow_id"); workflowIDStr != "" {
		if workflowID, err := uuid.Parse(workflowIDStr); err == nil {
			filter.WorkflowID = &workflowID
		}
	}

	// Trigger type filter
	if triggerType := r.URL.Query().Get("trigger_type"); triggerType != "" {
		filter.TriggerType = &triggerType
	}

	// Date range filters
	if startStr := r.URL.Query().Get("start_date"); startStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startStr); err == nil {
			filter.StartDate = &startTime
		}
	}
	if endStr := r.URL.Query().Get("end_date"); endStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endStr); err == nil {
			filter.EndDate = &endTime
		}
	}

	// Search query (searches in error messages)
	if q := r.URL.Query().Get("q"); q != "" {
		filter.SearchQuery = &q
	}

	executions, total, err := h.executionSvc.Search(r.Context(), filter, opts)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to search executions")
		return
	}

	response := []dto.ExecutionResponse{}
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

// BulkDelete deletes multiple executions
func (h *ExecutionHandler) BulkDelete(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "workspace context required")
		return
	}

	var req struct {
		ExecutionIDs []string `json:"execution_ids"`
		OlderThan    *string  `json:"older_than"` // RFC3339 format
		Status       *string  `json:"status"`     // Only delete specific status
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var deleted int64
	var err error

	if len(req.ExecutionIDs) > 0 {
		// Delete by IDs
		ids := make([]uuid.UUID, 0, len(req.ExecutionIDs))
		for _, idStr := range req.ExecutionIDs {
			if id, parseErr := uuid.Parse(idStr); parseErr == nil {
				ids = append(ids, id)
			}
		}
		deleted, err = h.executionSvc.DeleteByIDs(r.Context(), wsCtx.WorkspaceID, ids)
	} else if req.OlderThan != nil {
		// Delete by age
		cutoff, parseErr := time.Parse(time.RFC3339, *req.OlderThan)
		if parseErr != nil {
			dto.ErrorResponse(w, http.StatusBadRequest, "invalid older_than format (use RFC3339)")
			return
		}
		deleted, err = h.executionSvc.DeleteOlderThan(r.Context(), cutoff)
	} else {
		dto.ErrorResponse(w, http.StatusBadRequest, "provide execution_ids or older_than")
		return
	}

	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to delete executions")
		return
	}

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"deleted": deleted,
	})
}

// Stats returns execution statistics
func (h *ExecutionHandler) Stats(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "workspace context required")
		return
	}

	// Default to last 24 hours
	period := r.URL.Query().Get("period")
	var start, end time.Time
	end = time.Now()

	switch period {
	case "7d":
		start = end.AddDate(0, 0, -7)
	case "30d":
		start = end.AddDate(0, 0, -30)
	case "1h":
		start = end.Add(-time.Hour)
	default:
		start = end.AddDate(0, 0, -1) // Default 24h
	}

	stats, err := h.executionSvc.GetStats(r.Context(), wsCtx.WorkspaceID, start, end)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to get stats")
		return
	}

	dto.JSON(w, http.StatusOK, stats)
}
