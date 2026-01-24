package nodetypes

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/worker/nodes"
	"github.com/linkflow-ai/linkflow/internal/core/domain/sitesettings"
)

type ListNodeTypesHandler struct {
	registry         *nodes.Registry
	siteSettingsRepo sitesettings.Repository
}

func NewListNodeTypesHandler(registry *nodes.Registry, siteSettingsRepo sitesettings.Repository) *ListNodeTypesHandler {
	return &ListNodeTypesHandler{
		registry:         registry,
		siteSettingsRepo: siteSettingsRepo,
	}
}

func (h *ListNodeTypesHandler) Handle(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")

	// Get disabled nodes from site settings
	disabledNodes, err := h.siteSettingsRepo.GetDisabledNodes(r.Context())
	if err != nil {
		// Log error but continue with empty disabled list
		disabledNodes = []string{}
	}

	// Create a set for O(1) lookup
	disabledSet := make(map[string]bool)
	for _, n := range disabledNodes {
		disabledSet[n] = true
	}

	if category != "" {
		// Filter by category
		byCategory := h.registry.ListByCategory()
		if nodeTypes, ok := byCategory[category]; ok {
			// Filter out disabled nodes
			filtered := filterDisabledNodes(nodeTypes, disabledSet)
			common.Success(w, filtered)
			return
		}
		common.Success(w, []interface{}{})
		return
	}

	// Return all node types (filtered)
	metadata := h.registry.ListMetadata()
	filtered := filterDisabledNodes(metadata, disabledSet)
	common.Success(w, filtered)
}

// filterDisabledNodes removes disabled nodes from the list
func filterDisabledNodes(nodeTypes []nodes.NodeMetadata, disabledSet map[string]bool) []nodes.NodeMetadata {
	if len(disabledSet) == 0 {
		return nodeTypes
	}

	filtered := make([]nodes.NodeMetadata, 0, len(nodeTypes))
	for _, n := range nodeTypes {
		if !disabledSet[n.Type] {
			filtered = append(filtered, n)
		}
	}
	return filtered
}
