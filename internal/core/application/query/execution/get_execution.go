package execution

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// GetExecutionQuery represents the query to get an execution by ID
type GetExecutionQuery struct {
	ExecutionID uuid.UUID
}

// ListExecutionsQuery represents the query to list executions
type ListExecutionsQuery struct {
	WorkspaceID uuid.UUID
	WorkflowID  *uuid.UUID
	Status      *execution.Status
	TriggerType *string
	TriggeredBy *uuid.UUID
	DateFrom    *string
	DateTo      *string
	Page        int
	PageSize    int
}

// ListExecutionsResult contains the result of listing executions
type ListExecutionsResult struct {
	Executions []execution.Execution
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

// GetExecutionHandler handles getting executions
type GetExecutionHandler struct {
	executionRepo execution.Repository
}

// NewGetExecutionHandler creates a new handler
func NewGetExecutionHandler(executionRepo execution.Repository) *GetExecutionHandler {
	return &GetExecutionHandler{executionRepo: executionRepo}
}

// Handle executes the get execution query
func (h *GetExecutionHandler) Handle(ctx context.Context, q GetExecutionQuery) (*execution.Execution, error) {
	return h.executionRepo.FindByID(ctx, q.ExecutionID)
}

// ListExecutionsHandler handles listing executions
type ListExecutionsHandler struct {
	executionRepo execution.Repository
}

// NewListExecutionsHandler creates a new handler
func NewListExecutionsHandler(executionRepo execution.Repository) *ListExecutionsHandler {
	return &ListExecutionsHandler{executionRepo: executionRepo}
}

// Handle executes the list executions query
func (h *ListExecutionsHandler) Handle(ctx context.Context, q ListExecutionsQuery) (*ListExecutionsResult, error) {
	opts := execution.NewListOptions(q.Page, q.PageSize)
	opts.Status = q.Status
	opts.TriggerType = q.TriggerType
	opts.WorkflowID = q.WorkflowID
	opts.TriggeredBy = q.TriggeredBy

	executions, total, err := h.executionRepo.FindByWorkspaceID(ctx, q.WorkspaceID, opts)
	if err != nil {
		return nil, err
	}

	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = types.DefaultPageSize
	}
	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	return &ListExecutionsResult{
		Executions: executions,
		Total:      total,
		Page:       q.Page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetNodeExecutionsQuery represents the query to get node executions
type GetNodeExecutionsQuery struct {
	ExecutionID uuid.UUID
}

// GetNodeExecutionsHandler handles getting node executions
type GetNodeExecutionsHandler struct {
	nodeExecRepo execution.NodeExecutionRepository
}

// NewGetNodeExecutionsHandler creates a new handler
func NewGetNodeExecutionsHandler(nodeExecRepo execution.NodeExecutionRepository) *GetNodeExecutionsHandler {
	return &GetNodeExecutionsHandler{nodeExecRepo: nodeExecRepo}
}

// Handle executes the get node executions query
func (h *GetNodeExecutionsHandler) Handle(ctx context.Context, q GetNodeExecutionsQuery) ([]execution.NodeExecution, error) {
	return h.nodeExecRepo.FindByExecutionID(ctx, q.ExecutionID)
}
