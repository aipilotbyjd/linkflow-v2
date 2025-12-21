package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/worker/nodes"
)

type NodeTypeHandler struct {
	workflowSvc  *services.WorkflowService
	executionSvc *services.ExecutionService
}

func NewNodeTypeHandler(workflowSvc *services.WorkflowService, executionSvc *services.ExecutionService) *NodeTypeHandler {
	return &NodeTypeHandler{
		workflowSvc:  workflowSvc,
		executionSvc: executionSvc,
	}
}

// NodeTypeResponse represents a node type for the editor
type NodeTypeResponse struct {
	Type        string   `json:"type"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Icon        string   `json:"icon,omitempty"`
	Version     string   `json:"version,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Schema      *NodeSchema `json:"schema,omitempty"`
}

// NodeSchema defines input/output schema for a node
type NodeSchema struct {
	Inputs  []SchemaField `json:"inputs,omitempty"`
	Outputs []SchemaField `json:"outputs,omitempty"`
}

// SchemaField defines a field in node schema
type SchemaField struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Label       string      `json:"label"`
	Description string      `json:"description,omitempty"`
	Required    bool        `json:"required,omitempty"`
	Default     interface{} `json:"default,omitempty"`
	Options     []Option    `json:"options,omitempty"`
}

// Option for select/enum fields
type Option struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// ListNodeTypes returns all available node types for the workflow editor
func (h *NodeTypeHandler) ListNodeTypes(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")

	var metas []nodes.NodeMeta
	if category != "" {
		metas = nodes.ListByCategory(category)
	} else {
		metas = nodes.ListAll()
	}

	response := make([]NodeTypeResponse, len(metas))
	for i, meta := range metas {
		response[i] = NodeTypeResponse{
			Type:        meta.Type,
			Name:        meta.Name,
			Description: meta.Description,
			Category:    meta.Category,
			Icon:        meta.Icon,
			Version:     meta.Version,
			Tags:        meta.Tags,
			Schema:      getNodeSchema(meta.Type),
		}
	}

	dto.JSON(w, http.StatusOK, response)
}

// GetNodeType returns details for a specific node type
func (h *NodeTypeHandler) GetNodeType(w http.ResponseWriter, r *http.Request) {
	nodeType := chi.URLParam(r, "nodeType")

	meta, ok := nodes.GetMeta(nodeType)
	if !ok {
		dto.ErrorResponse(w, http.StatusNotFound, "node type not found")
		return
	}

	response := NodeTypeResponse{
		Type:        meta.Type,
		Name:        meta.Name,
		Description: meta.Description,
		Category:    meta.Category,
		Icon:        meta.Icon,
		Version:     meta.Version,
		Tags:        meta.Tags,
		Schema:      getNodeSchema(meta.Type),
	}

	dto.JSON(w, http.StatusOK, response)
}

// GetNodeCategories returns available node categories
func (h *NodeTypeHandler) GetNodeCategories(w http.ResponseWriter, r *http.Request) {
	metas := nodes.ListAll()

	categoryMap := make(map[string]int)
	for _, meta := range metas {
		categoryMap[meta.Category]++
	}

	categories := []map[string]interface{}{}
	categoryInfo := map[string]map[string]string{
		"trigger":     {"name": "Triggers", "description": "Start workflow execution", "icon": "play"},
		"action":      {"name": "Actions", "description": "Perform operations", "icon": "zap"},
		"logic":       {"name": "Logic", "description": "Control flow and conditions", "icon": "git-branch"},
		"integration": {"name": "Integrations", "description": "External service integrations", "icon": "plug"},
	}

	for cat, count := range categoryMap {
		info := categoryInfo[cat]
		if info == nil {
			info = map[string]string{"name": cat, "description": "", "icon": "box"}
		}
		categories = append(categories, map[string]interface{}{
			"id":          cat,
			"name":        info["name"],
			"description": info["description"],
			"icon":        info["icon"],
			"count":       count,
		})
	}

	dto.JSON(w, http.StatusOK, categories)
}

// ValidateWorkflow validates a workflow definition
func (h *NodeTypeHandler) ValidateWorkflow(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Nodes       models.JSONArray `json:"nodes"`
		Connections models.JSONArray `json:"connections"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	errors := validateWorkflowDefinition(req.Nodes, req.Connections)

	if len(errors) > 0 {
		dto.JSON(w, http.StatusOK, map[string]interface{}{
			"valid":  false,
			"errors": errors,
		})
		return
	}

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"valid":  true,
		"errors": []string{},
	})
}

// TestNode executes a single node for testing in the editor
func (h *NodeTypeHandler) TestNode(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if claims == nil || wsCtx == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "unauthorized")
		return
	}

	var req struct {
		NodeType   string                 `json:"node_type"`
		Parameters map[string]interface{} `json:"parameters"`
		Input      map[string]interface{} `json:"input"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	node := nodes.Get(req.NodeType)
	if node == nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "unknown node type: "+req.NodeType)
		return
	}

	// Build execution context
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	execCtx := &nodes.ExecutionContext{
		ExecutionID: uuid.New(),
		WorkflowID:  uuid.New(),
		WorkspaceID: wsCtx.WorkspaceID,
		NodeID:      "test",
		Input:       req.Input,
		Config:      req.Parameters,
		Variables:   make(map[string]interface{}),
	}

	start := time.Now()
	output, err := node.Execute(ctx, execCtx)
	duration := time.Since(start)

	if err != nil {
		dto.JSON(w, http.StatusOK, map[string]interface{}{
			"success":  false,
			"error":    err.Error(),
			"duration": duration.Milliseconds(),
		})
		return
	}

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"output":   output,
		"duration": duration.Milliseconds(),
	})
}

// validateWorkflowDefinition validates workflow nodes and connections
func validateWorkflowDefinition(nodesData models.JSONArray, connections models.JSONArray) []map[string]interface{} {
	var errors []map[string]interface{}

	if len(nodesData) == 0 {
		errors = append(errors, map[string]interface{}{
			"type":    "error",
			"message": "Workflow must have at least one node",
		})
		return errors
	}

	nodeIDs := make(map[string]bool)
	hasTrigger := false

	// Validate nodes
	for i, nodeData := range nodesData {
		node, ok := nodeData.(map[string]interface{})
		if !ok {
			errors = append(errors, map[string]interface{}{
				"type":    "error",
				"node":    i,
				"message": "Invalid node format",
			})
			continue
		}

		id, _ := node["id"].(string)
		nodeType, _ := node["type"].(string)

		if id == "" {
			errors = append(errors, map[string]interface{}{
				"type":    "error",
				"node":    i,
				"message": "Node missing id",
			})
		} else {
			if nodeIDs[id] {
				errors = append(errors, map[string]interface{}{
					"type":    "error",
					"node":    id,
					"message": "Duplicate node id",
				})
			}
			nodeIDs[id] = true
		}

		if nodeType == "" {
			errors = append(errors, map[string]interface{}{
				"type":    "error",
				"node":    id,
				"message": "Node missing type",
			})
		} else {
			// Check if node type exists
			if nodes.Get(nodeType) == nil {
				errors = append(errors, map[string]interface{}{
					"type":    "error",
					"node":    id,
					"message": "Unknown node type: " + nodeType,
				})
			}

			// Check for trigger
			if len(nodeType) > 8 && nodeType[:8] == "trigger." {
				hasTrigger = true
			}
		}
	}

	if !hasTrigger {
		errors = append(errors, map[string]interface{}{
			"type":    "warning",
			"message": "Workflow should have at least one trigger node",
		})
	}

	// Validate connections
	for i, connData := range connections {
		conn, ok := connData.(map[string]interface{})
		if !ok {
			errors = append(errors, map[string]interface{}{
				"type":       "error",
				"connection": i,
				"message":    "Invalid connection format",
			})
			continue
		}

		source, _ := conn["source"].(string)
		target, _ := conn["target"].(string)

		if source == "" || target == "" {
			errors = append(errors, map[string]interface{}{
				"type":       "error",
				"connection": i,
				"message":    "Connection missing source or target",
			})
			continue
		}

		if !nodeIDs[source] {
			errors = append(errors, map[string]interface{}{
				"type":       "error",
				"connection": i,
				"message":    "Connection source node not found: " + source,
			})
		}

		if !nodeIDs[target] {
			errors = append(errors, map[string]interface{}{
				"type":       "error",
				"connection": i,
				"message":    "Connection target node not found: " + target,
			})
		}
	}

	return errors
}

// getNodeSchema returns schema for a node type
func getNodeSchema(nodeType string) *NodeSchema {
	schemas := map[string]*NodeSchema{
		"trigger.manual": {
			Inputs: []SchemaField{},
			Outputs: []SchemaField{
				{Name: "triggered", Type: "boolean", Label: "Triggered"},
				{Name: "timestamp", Type: "string", Label: "Timestamp"},
			},
		},
		"trigger.webhook": {
			Inputs: []SchemaField{},
			Outputs: []SchemaField{
				{Name: "method", Type: "string", Label: "HTTP Method"},
				{Name: "headers", Type: "object", Label: "Headers"},
				{Name: "body", Type: "any", Label: "Body"},
				{Name: "query", Type: "object", Label: "Query Parameters"},
			},
		},
		"trigger.schedule": {
			Inputs: []SchemaField{
				{Name: "cron", Type: "string", Label: "Cron Expression", Description: "e.g., 0 9 * * 1-5"},
			},
			Outputs: []SchemaField{
				{Name: "scheduledTime", Type: "string", Label: "Scheduled Time"},
			},
		},
		"action.http": {
			Inputs: []SchemaField{
				{Name: "url", Type: "string", Label: "URL", Required: true},
				{Name: "method", Type: "select", Label: "Method", Default: "GET", Options: []Option{
					{Value: "GET", Label: "GET"},
					{Value: "POST", Label: "POST"},
					{Value: "PUT", Label: "PUT"},
					{Value: "PATCH", Label: "PATCH"},
					{Value: "DELETE", Label: "DELETE"},
				}},
				{Name: "headers", Type: "object", Label: "Headers"},
				{Name: "body", Type: "any", Label: "Body"},
			},
			Outputs: []SchemaField{
				{Name: "status", Type: "number", Label: "Status Code"},
				{Name: "headers", Type: "object", Label: "Response Headers"},
				{Name: "body", Type: "any", Label: "Response Body"},
				{Name: "json", Type: "object", Label: "JSON Response"},
			},
		},
		"action.set": {
			Inputs: []SchemaField{
				{Name: "values", Type: "object", Label: "Values to Set", Required: true},
			},
			Outputs: []SchemaField{
				{Name: "data", Type: "object", Label: "Set Values"},
			},
		},
		"action.code": {
			Inputs: []SchemaField{
				{Name: "language", Type: "select", Label: "Language", Default: "javascript", Options: []Option{
					{Value: "javascript", Label: "JavaScript"},
					{Value: "expr", Label: "Expression"},
				}},
				{Name: "code", Type: "code", Label: "Code", Required: true},
			},
			Outputs: []SchemaField{
				{Name: "result", Type: "any", Label: "Result"},
			},
		},
		"logic.condition": {
			Inputs: []SchemaField{
				{Name: "conditions", Type: "array", Label: "Conditions", Required: true},
			},
			Outputs: []SchemaField{
				{Name: "true", Type: "boolean", Label: "True Branch"},
				{Name: "false", Type: "boolean", Label: "False Branch"},
			},
		},
		"logic.loop": {
			Inputs: []SchemaField{
				{Name: "items", Type: "array", Label: "Items to Loop", Required: true},
				{Name: "batchSize", Type: "number", Label: "Batch Size", Default: 1},
			},
			Outputs: []SchemaField{
				{Name: "item", Type: "any", Label: "Current Item"},
				{Name: "index", Type: "number", Label: "Current Index"},
			},
		},
		"logic.switch": {
			Inputs: []SchemaField{
				{Name: "value", Type: "any", Label: "Value to Switch", Required: true},
				{Name: "cases", Type: "array", Label: "Cases"},
			},
			Outputs: []SchemaField{
				{Name: "matched", Type: "string", Label: "Matched Case"},
			},
		},
	}

	if schema, ok := schemas[nodeType]; ok {
		return schema
	}
	return nil
}
