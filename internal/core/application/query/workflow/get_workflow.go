package workflow

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// GetWorkflowQuery represents the query to get a workflow by ID
type GetWorkflowQuery struct {
	WorkflowID uuid.UUID
}

// ListWorkflowsQuery represents the query to list workflows
type ListWorkflowsQuery struct {
	WorkspaceID uuid.UUID
	Status      *workflow.Status
	FolderID    *uuid.UUID
	IsFavorite  *bool
	Search      string
	Tags        []string
	Page        int
	PageSize    int
}

// ListWorkflowsResult contains the result of listing workflows
type ListWorkflowsResult struct {
	Workflows  []workflow.Workflow
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

// GetWorkflowHandler handles getting workflows
type GetWorkflowHandler struct {
	workflowRepo workflow.Repository
}

// NewGetWorkflowHandler creates a new handler
func NewGetWorkflowHandler(workflowRepo workflow.Repository) *GetWorkflowHandler {
	return &GetWorkflowHandler{workflowRepo: workflowRepo}
}

// Handle executes the get workflow query
func (h *GetWorkflowHandler) Handle(ctx context.Context, q GetWorkflowQuery) (*workflow.Workflow, error) {
	return h.workflowRepo.FindByID(ctx, q.WorkflowID)
}

// ListWorkflowsHandler handles listing workflows
type ListWorkflowsHandler struct {
	workflowRepo workflow.Repository
}

// NewListWorkflowsHandler creates a new handler
func NewListWorkflowsHandler(workflowRepo workflow.Repository) *ListWorkflowsHandler {
	return &ListWorkflowsHandler{workflowRepo: workflowRepo}
}

// Handle executes the list workflows query
func (h *ListWorkflowsHandler) Handle(ctx context.Context, q ListWorkflowsQuery) (*ListWorkflowsResult, error) {
	opts := workflow.NewListOptions(q.Page, q.PageSize)
	opts.Status = q.Status
	opts.FolderID = q.FolderID
	opts.IsFavorite = q.IsFavorite
	opts.Search = q.Search
	opts.Tags = q.Tags

	workflows, total, err := h.workflowRepo.FindByWorkspaceID(ctx, q.WorkspaceID, opts)
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

	return &ListWorkflowsResult{
		Workflows:  workflows,
		Total:      total,
		Page:       q.Page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}
