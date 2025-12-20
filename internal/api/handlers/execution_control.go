package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/worker/processor"
	"github.com/redis/go-redis/v9"
)

// ExecutionControlHandler handles execution control operations
type ExecutionControlHandler struct {
	executionSvc *services.ExecutionService
	workflowSvc  *services.WorkflowService
	cancellation *processor.CancellationManager
	redis        *redis.Client
}

// NewExecutionControlHandler creates a new execution control handler
func NewExecutionControlHandler(
	executionSvc *services.ExecutionService,
	workflowSvc *services.WorkflowService,
	cancellation *processor.CancellationManager,
	redis *redis.Client,
) *ExecutionControlHandler {
	return &ExecutionControlHandler{
		executionSvc: executionSvc,
		workflowSvc:  workflowSvc,
		cancellation: cancellation,
		redis:        redis,
	}
}

// CancelExecutionRequest represents a cancellation request
type CancelExecutionRequest struct {
	Reason string `json:"reason"`
}

// CancelExecution cancels a running execution
func (h *ExecutionControlHandler) CancelExecution(w http.ResponseWriter, r *http.Request) {
	executionIDStr := chi.URLParam(r, "executionID")
	executionID, err := uuid.Parse(executionIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid execution ID")
		return
	}

	var req CancelExecutionRequest
	json.NewDecoder(r.Body).Decode(&req)

	if req.Reason == "" {
		req.Reason = "Cancelled by user"
	}

	// Get user from context
	requestedBy := "unknown"
	if wsCtx := middleware.GetWorkspaceFromContext(r.Context()); wsCtx != nil {
		requestedBy = wsCtx.WorkspaceID.String()
	}

	// Cancel the execution
	if h.cancellation != nil {
		if err := h.cancellation.Cancel(r.Context(), executionID, req.Reason, requestedBy); err != nil {
			dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	// Update execution status in database
	h.executionSvc.Fail(r.Context(), executionID, req.Reason, nil)

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"message":      "Execution cancelled",
		"execution_id": executionID,
		"reason":       req.Reason,
	})
}

// GetExecutionProgress returns progress for a running execution
func (h *ExecutionControlHandler) GetExecutionProgress(w http.ResponseWriter, r *http.Request) {
	executionIDStr := chi.URLParam(r, "executionID")
	executionID, err := uuid.Parse(executionIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid execution ID")
		return
	}

	progress, err := processor.GetProgressByID(r.Context(), h.redis, executionID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if progress == nil {
		// Try to get from database
		execution, err := h.executionSvc.GetByID(r.Context(), executionID)
		if err != nil {
			dto.ErrorResponse(w, http.StatusNotFound, "execution not found")
			return
		}

		dto.JSON(w, http.StatusOK, map[string]interface{}{
			"execution_id": executionID,
			"status":       execution.Status,
			"started_at":   execution.StartedAt,
			"completed_at": execution.CompletedAt,
		})
		return
	}

	dto.JSON(w, http.StatusOK, progress)
}

// PreviewWorkflowRequest represents a preview request
type PreviewWorkflowRequest struct {
	Input map[string]interface{} `json:"input"`
}

// PreviewWorkflow performs a dry-run validation of a workflow
func (h *ExecutionControlHandler) PreviewWorkflow(w http.ResponseWriter, r *http.Request) {
	workflowIDStr := chi.URLParam(r, "workflowID")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid workflow ID")
		return
	}

	var req PreviewWorkflowRequest
	json.NewDecoder(r.Body).Decode(&req)

	// Get workflow
	workflow, err := h.workflowSvc.GetByID(r.Context(), workflowID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusNotFound, "workflow not found")
		return
	}

	// Parse workflow definition
	workflowDef, err := processor.ParseWorkflow(workflow)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid workflow definition")
		return
	}

	// Create processor for preview
	proc := processor.New(processor.Config{})
	result, err := proc.Preview(r.Context(), workflowDef, req.Input)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, result)
}

// ValidateWorkflow validates a workflow's DAG
func (h *ExecutionControlHandler) ValidateWorkflow(w http.ResponseWriter, r *http.Request) {
	workflowIDStr := chi.URLParam(r, "workflowID")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid workflow ID")
		return
	}

	// Get workflow
	workflow, err := h.workflowSvc.GetByID(r.Context(), workflowID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusNotFound, "workflow not found")
		return
	}

	// Parse workflow definition
	workflowDef, err := processor.ParseWorkflow(workflow)
	if err != nil {
		dto.JSON(w, http.StatusOK, map[string]interface{}{
			"valid":  false,
			"errors": []string{err.Error()},
		})
		return
	}

	// Build DAG and validate
	dag := processor.BuildDAG(workflowDef)
	errors := dag.Validate()

	if len(errors) > 0 {
		errorStrings := make([]map[string]interface{}, len(errors))
		for i, e := range errors {
			errorStrings[i] = map[string]interface{}{
				"node_id": e.NodeID,
				"message": e.Message,
				"code":    e.Code,
			}
		}
		dto.JSON(w, http.StatusOK, map[string]interface{}{
			"valid":  false,
			"errors": errorStrings,
		})
		return
	}

	// Get execution order
	order, err := dag.TopologicalSort()
	if err != nil {
		dto.JSON(w, http.StatusOK, map[string]interface{}{
			"valid":  false,
			"errors": []string{err.Error()},
		})
		return
	}

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"valid":           true,
		"node_count":      dag.NodeCount(),
		"execution_order": order,
		"root_nodes":      dag.RootNodes,
		"leaf_nodes":      dag.LeafNodes,
	})
}

// GetActiveExecutions returns currently active executions
func (h *ExecutionControlHandler) GetActiveExecutions(w http.ResponseWriter, r *http.Request) {
	if h.cancellation == nil {
		dto.JSON(w, http.StatusOK, map[string]interface{}{
			"active": []string{},
			"count":  0,
		})
		return
	}

	activeIDs := h.cancellation.GetActiveExecutions()
	idStrings := make([]string, len(activeIDs))
	for i, id := range activeIDs {
		idStrings[i] = id.String()
	}

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"active": idStrings,
		"count":  len(idStrings),
	})
}

// WorkerStatsHandler handles worker statistics
type WorkerStatsHandler struct {
	cancellation *processor.CancellationManager
	redis        *redis.Client
}

// NewWorkerStatsHandler creates a new worker stats handler
func NewWorkerStatsHandler(cancellation *processor.CancellationManager, redis *redis.Client) *WorkerStatsHandler {
	return &WorkerStatsHandler{
		cancellation: cancellation,
		redis:        redis,
	}
}

// GetWorkerStats returns worker statistics
func (h *WorkerStatsHandler) GetWorkerStats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{
		"active_executions": 0,
	}

	if h.cancellation != nil {
		stats["active_executions"] = h.cancellation.ActiveCount()
	}

	dto.JSON(w, http.StatusOK, stats)
}

// GetWorkerHealth returns worker health status
func (h *WorkerStatsHandler) GetWorkerHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check Redis connection
	redisOk := true
	if err := h.redis.Ping(ctx).Err(); err != nil {
		redisOk = false
	}

	status := "healthy"
	if !redisOk {
		status = "degraded"
	}

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"status": status,
		"checks": map[string]interface{}{
			"redis": redisOk,
		},
	})
}
