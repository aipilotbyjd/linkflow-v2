package workspace

import (
	"net/http"
	"strconv"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	workspaceQuery "github.com/linkflow-ai/linkflow/internal/core/application/query/workspace"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// ListMembersHandler handles listing workspace members
type ListMembersHandler struct {
	handler *workspaceQuery.ListMembersHandler
}

// NewListMembersHandler creates a new handler
func NewListMembersHandler(handler *workspaceQuery.ListMembersHandler) *ListMembersHandler {
	return &ListMembersHandler{handler: handler}
}

// Handle handles the list members request
func (h *ListMembersHandler) Handle(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		common.BadRequest(w, "workspace context required")
		return
	}

	query := r.URL.Query()
	page := 1
	if p := query.Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	pageSize := types.DefaultPageSize
	if ps := query.Get("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}

	result, err := h.handler.Handle(r.Context(), workspaceQuery.ListMembersQuery{
		WorkspaceID: wsCtx.WorkspaceID,
		Page:        page,
		PageSize:    pageSize,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	members := make([]MemberResponse, len(result.Members))
	for i, m := range result.Members {
		members[i] = ToMemberResponse(&m)
	}

	common.List(w, members, types.PageResponse{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalItems: result.Total,
		TotalPages: result.TotalPages,
		HasMore:    result.Page < result.TotalPages,
	})
}
