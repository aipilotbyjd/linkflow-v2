package workflow

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/adapters/worker/nodes"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/validation"
)

type ValidateWorkflowHandler struct {
	nodeRegistry *nodes.Registry
}

func NewValidateWorkflowHandler(nodeRegistry *nodes.Registry) *ValidateWorkflowHandler {
	return &ValidateWorkflowHandler{nodeRegistry: nodeRegistry}
}

func (h *ValidateWorkflowHandler) Handle(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		common.BadRequest(w, "workspace context required")
		return
	}

	var req ValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}

	if errors := validation.Validate(req); len(errors) > 0 {
		details := make([]common.ValidationDetail, len(errors))
		for i, e := range errors {
			details[i] = common.ValidationDetail{Field: e.Field, Message: e.Message}
		}
		common.ValidationErrors(w, details)
		return
	}

	result := ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []ValidationError{},
	}

	// Check for empty workflow
	if len(req.Nodes) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Message: "Workflow must have at least one node",
			Code:    "EMPTY_WORKFLOW",
		})
	}

	// Check for unique node IDs
	nodeIDs := make(map[string]bool)
	for _, node := range req.Nodes {
		if nodeIDs[node.ID] {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				NodeID:  node.ID,
				Message: "Duplicate node ID",
				Code:    "DUPLICATE_NODE_ID",
			})
		}
		nodeIDs[node.ID] = true
	}

	// Check for valid node types
	hasTrigger := false
	for _, node := range req.Nodes {
		if !h.nodeRegistry.Has(node.Type) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				NodeID:  node.ID,
				Field:   "type",
				Message: "Unknown node type: " + node.Type,
				Code:    "UNKNOWN_NODE_TYPE",
			})
		}

		// Check for trigger node
		if len(node.Type) > 8 && node.Type[:8] == "trigger." {
			hasTrigger = true
		}
	}

	if !hasTrigger && len(req.Nodes) > 0 {
		result.Warnings = append(result.Warnings, ValidationError{
			Message: "Workflow has no trigger node",
			Code:    "NO_TRIGGER",
		})
	}

	// Validate connections
	for _, conn := range req.Connections {
		if !nodeIDs[conn.Source] {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "source",
				Message: "Connection source node not found: " + conn.Source,
				Code:    "INVALID_CONNECTION",
			})
		}
		if !nodeIDs[conn.Target] {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "target",
				Message: "Connection target node not found: " + conn.Target,
				Code:    "INVALID_CONNECTION",
			})
		}
	}

	// Check for orphan nodes (no connections)
	connectedNodes := make(map[string]bool)
	for _, conn := range req.Connections {
		connectedNodes[conn.Source] = true
		connectedNodes[conn.Target] = true
	}
	for _, node := range req.Nodes {
		if !connectedNodes[node.ID] && len(req.Nodes) > 1 {
			result.Warnings = append(result.Warnings, ValidationError{
				NodeID:  node.ID,
				Message: "Node is not connected to any other node",
				Code:    "ORPHAN_NODE",
			})
		}
	}

	common.Success(w, result)
}
