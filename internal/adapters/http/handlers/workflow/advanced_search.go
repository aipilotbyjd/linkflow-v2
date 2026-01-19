package workflow

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// AdvancedSearchRequest represents the advanced search request body
type AdvancedSearchRequest struct {
	// Text search
	Query    string   `json:"query,omitempty"`
	SearchIn []string `json:"search_in,omitempty"` // name, description, tags, nodes

	// Filters
	Status    []string `json:"status,omitempty"`     // draft, active, inactive, archived
	Tags      []string `json:"tags,omitempty"`       // Match any
	TagsAll   []string `json:"tags_all,omitempty"`   // Match all
	NodeTypes []string `json:"node_types,omitempty"` // Filter by node types used
	Category  string   `json:"category,omitempty"`
	Favorite  *bool    `json:"favorite,omitempty"`
	FolderID  string   `json:"folder_id,omitempty"`
	CreatedBy string   `json:"created_by,omitempty"`

	// Date filters (Unix timestamps)
	CreatedAfter   *int64 `json:"created_after,omitempty"`
	CreatedBefore  *int64 `json:"created_before,omitempty"`
	UpdatedAfter   *int64 `json:"updated_after,omitempty"`
	UpdatedBefore  *int64 `json:"updated_before,omitempty"`
	ExecutedAfter  *int64 `json:"executed_after,omitempty"`
	ExecutedBefore *int64 `json:"executed_before,omitempty"`

	// Execution filters
	MinExecutions *int  `json:"min_executions,omitempty"`
	MaxExecutions *int  `json:"max_executions,omitempty"`
	HasErrors     *bool `json:"has_errors,omitempty"`

	// Sorting
	SortBy    string `json:"sort_by,omitempty"`    // name, created_at, updated_at, execution_count, last_executed_at
	SortOrder string `json:"sort_order,omitempty"` // asc, desc

	// Pagination
	Page     int `json:"page,omitempty"`
	PageSize int `json:"page_size,omitempty"`
}

// AdvancedSearchHandler handles advanced workflow search
type AdvancedSearchHandler struct {
	workflowRepo workflow.SearchRepository
}

func NewAdvancedSearchHandler(workflowRepo workflow.SearchRepository) *AdvancedSearchHandler {
	return &AdvancedSearchHandler{workflowRepo: workflowRepo}
}

func (h *AdvancedSearchHandler) Handle(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		common.BadRequest(w, "workspace context required")
		return
	}

	var req AdvancedSearchRequest

	// Support both GET with query params and POST with body
	if r.Method == http.MethodPost {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.BadRequest(w, "invalid request body")
			return
		}
	} else {
		// Parse from query params
		req = parseAdvancedSearchFromQuery(r)
	}

	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}
	if len(req.SearchIn) == 0 {
		req.SearchIn = []string{"name", "description"}
	}

	// Build options
	opts := workflow.NewAdvancedSearchOptions(req.Page, req.PageSize)
	opts.Query = req.Query
	opts.SearchIn = req.SearchIn
	opts.Category = req.Category
	opts.IsFavorite = req.Favorite
	opts.SortBy = req.SortBy
	opts.SortOrder = req.SortOrder

	// Convert status strings to Status type
	for _, s := range req.Status {
		opts.Status = append(opts.Status, workflow.Status(s))
	}

	opts.Tags = req.Tags
	opts.TagsAll = req.TagsAll
	opts.NodeTypes = req.NodeTypes

	// Parse UUIDs
	if req.FolderID != "" {
		if id, err := uuid.Parse(req.FolderID); err == nil {
			opts.FolderID = &id
		}
	}
	if req.CreatedBy != "" {
		if id, err := uuid.Parse(req.CreatedBy); err == nil {
			opts.CreatedBy = &id
		}
	}

	// Date filters
	opts.CreatedAfter = req.CreatedAfter
	opts.CreatedBefore = req.CreatedBefore
	opts.UpdatedAfter = req.UpdatedAfter
	opts.UpdatedBefore = req.UpdatedBefore
	opts.ExecutedAfter = req.ExecutedAfter
	opts.ExecutedBefore = req.ExecutedBefore

	// Execution filters
	opts.MinExecutions = req.MinExecutions
	opts.MaxExecutions = req.MaxExecutions
	opts.HasErrors = req.HasErrors

	// Execute search
	workflows, total, err := h.workflowRepo.AdvancedSearch(r.Context(), wsCtx.WorkspaceID, opts)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	// Build response
	var responses []WorkflowResponse
	for _, wf := range workflows {
		responses = append(responses, ToWorkflowResponse(&wf))
	}

	common.List(w, responses, types.PageResponse{
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalItems: total,
		TotalPages: int((total + int64(req.PageSize) - 1) / int64(req.PageSize)),
		HasMore:    int64(req.Page*req.PageSize) < total,
	})
}

// parseAdvancedSearchFromQuery parses search params from URL query
func parseAdvancedSearchFromQuery(r *http.Request) AdvancedSearchRequest {
	q := r.URL.Query()

	req := AdvancedSearchRequest{
		Query:     q.Get("q"),
		Category:  q.Get("category"),
		FolderID:  q.Get("folder_id"),
		CreatedBy: q.Get("created_by"),
		SortBy:    q.Get("sort_by"),
		SortOrder: q.Get("sort_order"),
	}

	// Parse search_in
	if searchIn := q.Get("search_in"); searchIn != "" {
		req.SearchIn = strings.Split(searchIn, ",")
	}

	// Parse status
	if status := q.Get("status"); status != "" {
		req.Status = strings.Split(status, ",")
	}

	// Parse tags
	if tags := q.Get("tags"); tags != "" {
		req.Tags = strings.Split(tags, ",")
	}
	if tagsAll := q.Get("tags_all"); tagsAll != "" {
		req.TagsAll = strings.Split(tagsAll, ",")
	}

	// Parse node_types
	if nodeTypes := q.Get("node_types"); nodeTypes != "" {
		req.NodeTypes = strings.Split(nodeTypes, ",")
	}

	// Parse favorite
	if fav := q.Get("favorite"); fav != "" {
		b := fav == "true" || fav == "1"
		req.Favorite = &b
	}

	// Parse dates
	if v := q.Get("created_after"); v != "" {
		if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
			req.CreatedAfter = &ts
		}
	}
	if v := q.Get("created_before"); v != "" {
		if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
			req.CreatedBefore = &ts
		}
	}
	if v := q.Get("updated_after"); v != "" {
		if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
			req.UpdatedAfter = &ts
		}
	}
	if v := q.Get("updated_before"); v != "" {
		if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
			req.UpdatedBefore = &ts
		}
	}
	if v := q.Get("executed_after"); v != "" {
		if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
			req.ExecutedAfter = &ts
		}
	}
	if v := q.Get("executed_before"); v != "" {
		if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
			req.ExecutedBefore = &ts
		}
	}

	// Parse execution filters
	if v := q.Get("min_executions"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			req.MinExecutions = &n
		}
	}
	if v := q.Get("max_executions"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			req.MaxExecutions = &n
		}
	}
	if v := q.Get("has_errors"); v != "" {
		b := v == "true" || v == "1"
		req.HasErrors = &b
	}

	// Parse pagination
	req.Page, _ = strconv.Atoi(q.Get("page"))
	req.PageSize, _ = strconv.Atoi(q.Get("page_size"))

	return req
}
