package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/worker/processor"
	"github.com/redis/go-redis/v9"
)

// ExecutionControlHandler handles execution control operations
type ExecutionControlHandler struct {
	executionSvc  *services.ExecutionService
	workflowSvc   *services.WorkflowService
	cancellation  *processor.CancellationManager
	redis         *redis.Client
}

// NewExecutionControlHandler creates a new execution control handler
func NewExecutionControlHandler(
	executionSvc *services.ExecutionService,
	workflowSvc *services.WorkflowService,
	cancellation *processor.CancellationManager,
	redis *redis.Client,
) *ExecutionControlHandler {
	return &ExecutionControlHandler{
		executionSvc:  executionSvc,
		workflowSvc:   workflowSvc,
		cancellation:  cancellation,
		redis:         redis,
	}
}

// CancelExecutionRequest represents a cancellation request
type CancelExecutionRequest struct {
	Reason string `json:"reason"`
}

// CancelExecution cancels a running execution
// @Summary Cancel execution
// @Tags executions
// @Param id path string true "Execution ID"
// @Param body body CancelExecutionRequest false "Cancellation details"
// @Success 200 {object} map[string]interface{}
// @Router /executions/{id}/cancel [post]
func (h *ExecutionControlHandler) CancelExecution(c *gin.Context) {
	executionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid execution ID"})
		return
	}

	var req CancelExecutionRequest
	c.ShouldBindJSON(&req)

	if req.Reason == "" {
		req.Reason = "Cancelled by user"
	}

	// Get user from context
	requestedBy := "unknown"
	if userID, exists := c.Get("user_id"); exists {
		requestedBy = userID.(string)
	}

	// Cancel the execution
	if err := h.cancellation.Cancel(c.Request.Context(), executionID, req.Reason, requestedBy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update execution status in database
	h.executionSvc.Fail(c.Request.Context(), executionID, req.Reason, nil)

	c.JSON(http.StatusOK, gin.H{
		"message":      "Execution cancelled",
		"execution_id": executionID,
		"reason":       req.Reason,
	})
}

// GetExecutionProgress returns progress for a running execution
// @Summary Get execution progress
// @Tags executions
// @Param id path string true "Execution ID"
// @Success 200 {object} map[string]interface{}
// @Router /executions/{id}/progress [get]
func (h *ExecutionControlHandler) GetExecutionProgress(c *gin.Context) {
	executionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid execution ID"})
		return
	}

	progress, err := processor.GetProgressByID(c.Request.Context(), h.redis, executionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if progress == nil {
		// Try to get from database
		execution, err := h.executionSvc.GetByID(c.Request.Context(), executionID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Execution not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"execution_id": executionID,
			"status":       execution.Status,
			"started_at":   execution.StartedAt,
			"completed_at": execution.CompletedAt,
		})
		return
	}

	c.JSON(http.StatusOK, progress)
}

// PreviewWorkflowRequest represents a preview request
type PreviewWorkflowRequest struct {
	Input map[string]interface{} `json:"input"`
}

// PreviewWorkflow performs a dry-run validation of a workflow
// @Summary Preview workflow execution
// @Tags workflows
// @Param id path string true "Workflow ID"
// @Param body body PreviewWorkflowRequest false "Preview input"
// @Success 200 {object} processor.PreviewResult
// @Router /workflows/{id}/preview [post]
func (h *ExecutionControlHandler) PreviewWorkflow(c *gin.Context) {
	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workflow ID"})
		return
	}

	var req PreviewWorkflowRequest
	c.ShouldBindJSON(&req)

	// Get workflow
	workflow, err := h.workflowSvc.GetByID(c.Request.Context(), workflowID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	// Parse workflow definition
	workflowDef, err := processor.ParseWorkflow(workflow)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workflow definition"})
		return
	}

	// Create processor for preview
	proc := processor.New(processor.Config{})
	result, err := proc.Preview(c.Request.Context(), workflowDef, req.Input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ValidateWorkflow validates a workflow's DAG
// @Summary Validate workflow
// @Tags workflows
// @Param id path string true "Workflow ID"
// @Success 200 {object} map[string]interface{}
// @Router /workflows/{id}/validate [get]
func (h *ExecutionControlHandler) ValidateWorkflow(c *gin.Context) {
	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workflow ID"})
		return
	}

	// Get workflow
	workflow, err := h.workflowSvc.GetByID(c.Request.Context(), workflowID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	// Parse workflow definition
	workflowDef, err := processor.ParseWorkflow(workflow)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
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
		c.JSON(http.StatusOK, gin.H{
			"valid":  false,
			"errors": errorStrings,
		})
		return
	}

	// Get execution order
	order, err := dag.TopologicalSort()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"valid":  false,
			"errors": []string{err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":           true,
		"node_count":      dag.NodeCount(),
		"execution_order": order,
		"root_nodes":      dag.RootNodes,
		"leaf_nodes":      dag.LeafNodes,
	})
}

// GetActiveExecutions returns currently active executions
// @Summary Get active executions
// @Tags executions
// @Success 200 {object} map[string]interface{}
// @Router /executions/active [get]
func (h *ExecutionControlHandler) GetActiveExecutions(c *gin.Context) {
	if h.cancellation == nil {
		c.JSON(http.StatusOK, gin.H{
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

	c.JSON(http.StatusOK, gin.H{
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
// @Summary Get worker stats
// @Tags worker
// @Success 200 {object} map[string]interface{}
// @Router /worker/stats [get]
func (h *WorkerStatsHandler) GetWorkerStats(c *gin.Context) {
	stats := gin.H{
		"active_executions": 0,
	}

	if h.cancellation != nil {
		stats["active_executions"] = h.cancellation.ActiveCount()
	}

	c.JSON(http.StatusOK, stats)
}

// GetWorkerHealth returns worker health status
// @Summary Get worker health
// @Tags worker
// @Success 200 {object} map[string]interface{}
// @Router /worker/health [get]
func (h *WorkerStatsHandler) GetWorkerHealth(c *gin.Context) {
	ctx := c.Request.Context()

	// Check Redis connection
	redisOk := true
	if err := h.redis.Ping(ctx).Err(); err != nil {
		redisOk = false
	}

	status := "healthy"
	if !redisOk {
		status = "degraded"
	}

	c.JSON(http.StatusOK, gin.H{
		"status": status,
		"checks": gin.H{
			"redis": redisOk,
		},
	})
}
