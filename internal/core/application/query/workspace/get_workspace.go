package workspace

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workspace"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// GetWorkspaceQuery represents the query to get a workspace by ID
type GetWorkspaceQuery struct {
	WorkspaceID uuid.UUID
}

// GetWorkspaceBySlugQuery represents the query to get a workspace by slug
type GetWorkspaceBySlugQuery struct {
	Slug string
}

// ListWorkspacesQuery represents the query to list workspaces for a user
type ListWorkspacesQuery struct {
	UserID uuid.UUID
}

// ListMembersQuery represents the query to list workspace members
type ListMembersQuery struct {
	WorkspaceID uuid.UUID
	Page        int
	PageSize    int
}

// ListMembersResult contains the result of listing members
type ListMembersResult struct {
	Members    []workspace.Member
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

// GetWorkspaceHandler handles getting workspaces
type GetWorkspaceHandler struct {
	workspaceRepo workspace.Repository
}

// NewGetWorkspaceHandler creates a new handler
func NewGetWorkspaceHandler(workspaceRepo workspace.Repository) *GetWorkspaceHandler {
	return &GetWorkspaceHandler{workspaceRepo: workspaceRepo}
}

// Handle executes the get workspace query
func (h *GetWorkspaceHandler) Handle(ctx context.Context, q GetWorkspaceQuery) (*workspace.Workspace, error) {
	return h.workspaceRepo.FindByID(ctx, q.WorkspaceID)
}

// HandleBySlug gets a workspace by slug
func (h *GetWorkspaceHandler) HandleBySlug(ctx context.Context, q GetWorkspaceBySlugQuery) (*workspace.Workspace, error) {
	return h.workspaceRepo.FindBySlug(ctx, q.Slug)
}

// ListWorkspacesHandler handles listing workspaces
type ListWorkspacesHandler struct {
	memberRepo workspace.MemberRepository
}

// NewListWorkspacesHandler creates a new handler
func NewListWorkspacesHandler(memberRepo workspace.MemberRepository) *ListWorkspacesHandler {
	return &ListWorkspacesHandler{memberRepo: memberRepo}
}

// Handle executes the list workspaces query
func (h *ListWorkspacesHandler) Handle(ctx context.Context, q ListWorkspacesQuery) ([]workspace.Workspace, error) {
	return h.memberRepo.FindWorkspacesByUserID(ctx, q.UserID)
}

// ListMembersHandler handles listing workspace members
type ListMembersHandler struct {
	memberRepo workspace.MemberRepository
}

// NewListMembersHandler creates a new handler
func NewListMembersHandler(memberRepo workspace.MemberRepository) *ListMembersHandler {
	return &ListMembersHandler{memberRepo: memberRepo}
}

// Handle executes the list members query
func (h *ListMembersHandler) Handle(ctx context.Context, q ListMembersQuery) (*ListMembersResult, error) {
	opts := types.NewListOptions(q.Page, q.PageSize)

	members, total, err := h.memberRepo.FindByWorkspaceID(ctx, q.WorkspaceID, opts)
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

	return &ListMembersResult{
		Members:    members,
		Total:      total,
		Page:       q.Page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}
