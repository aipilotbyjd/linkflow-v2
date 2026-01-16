package workflow

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// GetVersionsQuery represents the query to get workflow versions
type GetVersionsQuery struct {
	WorkflowID uuid.UUID
	Page       int
	PageSize   int
}

// GetVersionsResult contains the result of getting versions
type GetVersionsResult struct {
	Versions   []workflow.Version
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

// GetVersionQuery represents the query to get a specific version
type GetVersionQuery struct {
	WorkflowID uuid.UUID
	Version    int
}

// GetVersionsHandler handles getting workflow versions
type GetVersionsHandler struct {
	versionRepo workflow.VersionRepository
}

// NewGetVersionsHandler creates a new handler
func NewGetVersionsHandler(versionRepo workflow.VersionRepository) *GetVersionsHandler {
	return &GetVersionsHandler{versionRepo: versionRepo}
}

// Handle executes the get versions query
func (h *GetVersionsHandler) Handle(ctx context.Context, q GetVersionsQuery) (*GetVersionsResult, error) {
	opts := types.NewListOptions(q.Page, q.PageSize)

	versions, total, err := h.versionRepo.FindByWorkflowID(ctx, q.WorkflowID, opts)
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

	return &GetVersionsResult{
		Versions:   versions,
		Total:      total,
		Page:       q.Page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// HandleSingle gets a specific version
func (h *GetVersionsHandler) HandleSingle(ctx context.Context, q GetVersionQuery) (*workflow.Version, error) {
	return h.versionRepo.FindByWorkflowAndVersion(ctx, q.WorkflowID, q.Version)
}
