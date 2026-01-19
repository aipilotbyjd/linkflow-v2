package workflow

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
)

// SearchFiltersHandler returns available search filters for the workspace
type SearchFiltersHandler struct {
	workflowRepo workflow.SearchRepository
}

func NewSearchFiltersHandler(workflowRepo workflow.SearchRepository) *SearchFiltersHandler {
	return &SearchFiltersHandler{workflowRepo: workflowRepo}
}

func (h *SearchFiltersHandler) Handle(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		common.BadRequest(w, "workspace context required")
		return
	}

	// Get all available filters in parallel
	type result struct {
		tags       []string
		nodeTypes  []string
		categories []string
		err        error
	}

	ch := make(chan result, 3)

	// Get tags
	go func() {
		tags, err := h.workflowRepo.GetTagsInWorkspace(r.Context(), wsCtx.WorkspaceID)
		ch <- result{tags: tags, err: err}
	}()

	// Get node types
	go func() {
		nodeTypes, err := h.workflowRepo.GetNodeTypesInWorkspace(r.Context(), wsCtx.WorkspaceID)
		ch <- result{nodeTypes: nodeTypes, err: err}
	}()

	// Get categories
	go func() {
		categories, err := h.workflowRepo.GetCategoriesInWorkspace(r.Context(), wsCtx.WorkspaceID)
		ch <- result{categories: categories, err: err}
	}()

	// Collect results
	var tags, nodeTypes, categories []string
	for i := 0; i < 3; i++ {
		res := <-ch
		if res.err != nil {
			// Log error but continue
			continue
		}
		if res.tags != nil {
			tags = res.tags
		}
		if res.nodeTypes != nil {
			nodeTypes = res.nodeTypes
		}
		if res.categories != nil {
			categories = res.categories
		}
	}

	// Ensure non-nil slices for JSON
	if tags == nil {
		tags = []string{}
	}
	if nodeTypes == nil {
		nodeTypes = []string{}
	}
	if categories == nil {
		categories = []string{}
	}

	common.Success(w, map[string]interface{}{
		"tags":       tags,
		"node_types": nodeTypes,
		"categories": categories,
		"statuses": []string{
			string(workflow.StatusDraft),
			string(workflow.StatusActive),
			string(workflow.StatusInactive),
			string(workflow.StatusArchived),
		},
		"sort_options": []map[string]string{
			{"value": "name", "label": "Name"},
			{"value": "created_at", "label": "Created Date"},
			{"value": "updated_at", "label": "Updated Date"},
			{"value": "execution_count", "label": "Execution Count"},
			{"value": "last_executed_at", "label": "Last Executed"},
		},
		"search_fields": []map[string]string{
			{"value": "name", "label": "Name"},
			{"value": "description", "label": "Description"},
			{"value": "tags", "label": "Tags"},
			{"value": "nodes", "label": "Node Content"},
		},
	})
}
