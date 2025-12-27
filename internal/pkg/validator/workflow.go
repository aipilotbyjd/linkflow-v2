package validator

import (
	"fmt"
	"strings"
)

// WorkflowNode represents a node in the workflow for validation
type WorkflowNode struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"parameters"`
}

// WorkflowConnection represents a connection between nodes
type WorkflowConnection struct {
	ID           string `json:"id"`
	SourceNodeID string `json:"source_node_id"`
	SourceHandle string `json:"source_handle"`
	TargetNodeID string `json:"target_node_id"`
	TargetHandle string `json:"target_handle"`
}

// WorkflowValidationError represents a validation error with context
type WorkflowValidationError struct {
	Field   string `json:"field"`
	NodeID  string `json:"node_id,omitempty"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e WorkflowValidationError) Error() string {
	if e.NodeID != "" {
		return fmt.Sprintf("[%s] %s: %s", e.NodeID, e.Code, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// WorkflowValidationResult holds all validation errors
type WorkflowValidationResult struct {
	Valid  bool                      `json:"valid"`
	Errors []WorkflowValidationError `json:"errors,omitempty"`
}

func (r *WorkflowValidationResult) AddError(err WorkflowValidationError) {
	r.Valid = false
	r.Errors = append(r.Errors, err)
}

func (r *WorkflowValidationResult) Error() string {
	if r.Valid {
		return ""
	}
	var msgs []string
	for _, e := range r.Errors {
		msgs = append(msgs, e.Error())
	}
	return strings.Join(msgs, "; ")
}

// NodeTypeChecker is a function that checks if a node type is valid
type NodeTypeChecker func(nodeType string) bool

// WorkflowValidator validates workflow structure
type WorkflowValidator struct {
	nodeTypeChecker NodeTypeChecker
}

// NewWorkflowValidator creates a new workflow validator
func NewWorkflowValidator(checker NodeTypeChecker) *WorkflowValidator {
	return &WorkflowValidator{
		nodeTypeChecker: checker,
	}
}

// Validate performs full workflow validation
func (v *WorkflowValidator) Validate(nodes []WorkflowNode, connections []WorkflowConnection) *WorkflowValidationResult {
	result := &WorkflowValidationResult{Valid: true}

	// Build node map for quick lookup
	nodeMap := make(map[string]*WorkflowNode)
	nodeIDs := make(map[string]bool)

	// 1. Validate nodes
	v.validateNodes(nodes, nodeMap, nodeIDs, result)

	// 2. Validate connections
	v.validateConnections(connections, nodeIDs, result)

	// 3. Validate graph structure
	v.validateGraphStructure(nodes, connections, nodeMap, result)

	return result
}

// validateNodes checks individual node validity
func (v *WorkflowValidator) validateNodes(nodes []WorkflowNode, nodeMap map[string]*WorkflowNode, nodeIDs map[string]bool, result *WorkflowValidationResult) {
	if len(nodes) == 0 {
		result.AddError(WorkflowValidationError{
			Field:   "nodes",
			Code:    "EMPTY_WORKFLOW",
			Message: "Workflow must have at least one node",
		})
		return
	}

	for i, node := range nodes {
		// Validate node ID
		if node.ID == "" {
			result.AddError(WorkflowValidationError{
				Field:   fmt.Sprintf("nodes[%d].id", i),
				Code:    "MISSING_NODE_ID",
				Message: "Node ID is required",
			})
			continue
		}

		// Check for duplicate IDs
		if nodeIDs[node.ID] {
			result.AddError(WorkflowValidationError{
				Field:   fmt.Sprintf("nodes[%d].id", i),
				NodeID:  node.ID,
				Code:    "DUPLICATE_NODE_ID",
				Message: fmt.Sprintf("Duplicate node ID: %s", node.ID),
			})
			continue
		}
		nodeIDs[node.ID] = true
		nodeMap[node.ID] = &nodes[i]

		// Validate node type
		if node.Type == "" {
			result.AddError(WorkflowValidationError{
				Field:  fmt.Sprintf("nodes[%d].type", i),
				NodeID: node.ID,
				Code:   "MISSING_NODE_TYPE",
				Message: "Node type is required",
			})
		} else if v.nodeTypeChecker != nil && !v.nodeTypeChecker(node.Type) {
			result.AddError(WorkflowValidationError{
				Field:   fmt.Sprintf("nodes[%d].type", i),
				NodeID:  node.ID,
				Code:    "INVALID_NODE_TYPE",
				Message: fmt.Sprintf("Unknown node type: %s", node.Type),
			})
		}

		// Validate node name (optional but recommended)
		if node.Name == "" {
			result.AddError(WorkflowValidationError{
				Field:   fmt.Sprintf("nodes[%d].name", i),
				NodeID:  node.ID,
				Code:    "MISSING_NODE_NAME",
				Message: "Node name is recommended",
			})
		}
	}
}

// validateConnections checks connection validity
func (v *WorkflowValidator) validateConnections(connections []WorkflowConnection, nodeIDs map[string]bool, result *WorkflowValidationResult) {
	connectionIDs := make(map[string]bool)
	connectionPairs := make(map[string]bool)

	for i, conn := range connections {
		// Validate connection ID
		if conn.ID == "" {
			result.AddError(WorkflowValidationError{
				Field:   fmt.Sprintf("connections[%d].id", i),
				Code:    "MISSING_CONNECTION_ID",
				Message: "Connection ID is required",
			})
		} else if connectionIDs[conn.ID] {
			result.AddError(WorkflowValidationError{
				Field:   fmt.Sprintf("connections[%d].id", i),
				Code:    "DUPLICATE_CONNECTION_ID",
				Message: fmt.Sprintf("Duplicate connection ID: %s", conn.ID),
			})
		} else {
			connectionIDs[conn.ID] = true
		}

		// Validate source node exists
		if conn.SourceNodeID == "" {
			result.AddError(WorkflowValidationError{
				Field:   fmt.Sprintf("connections[%d].source_node_id", i),
				Code:    "MISSING_SOURCE_NODE",
				Message: "Source node ID is required",
			})
		} else if !nodeIDs[conn.SourceNodeID] {
			result.AddError(WorkflowValidationError{
				Field:   fmt.Sprintf("connections[%d].source_node_id", i),
				Code:    "INVALID_SOURCE_NODE",
				Message: fmt.Sprintf("Source node not found: %s", conn.SourceNodeID),
			})
		}

		// Validate target node exists
		if conn.TargetNodeID == "" {
			result.AddError(WorkflowValidationError{
				Field:   fmt.Sprintf("connections[%d].target_node_id", i),
				Code:    "MISSING_TARGET_NODE",
				Message: "Target node ID is required",
			})
		} else if !nodeIDs[conn.TargetNodeID] {
			result.AddError(WorkflowValidationError{
				Field:   fmt.Sprintf("connections[%d].target_node_id", i),
				Code:    "INVALID_TARGET_NODE",
				Message: fmt.Sprintf("Target node not found: %s", conn.TargetNodeID),
			})
		}

		// Check for self-loops
		if conn.SourceNodeID != "" && conn.SourceNodeID == conn.TargetNodeID {
			result.AddError(WorkflowValidationError{
				Field:   fmt.Sprintf("connections[%d]", i),
				Code:    "SELF_LOOP",
				Message: fmt.Sprintf("Node cannot connect to itself: %s", conn.SourceNodeID),
			})
		}

		// Check for duplicate connections (same source-target pair with same handles)
		pairKey := fmt.Sprintf("%s:%s->%s:%s", conn.SourceNodeID, conn.SourceHandle, conn.TargetNodeID, conn.TargetHandle)
		if connectionPairs[pairKey] {
			result.AddError(WorkflowValidationError{
				Field:   fmt.Sprintf("connections[%d]", i),
				Code:    "DUPLICATE_CONNECTION",
				Message: fmt.Sprintf("Duplicate connection from %s to %s", conn.SourceNodeID, conn.TargetNodeID),
			})
		} else {
			connectionPairs[pairKey] = true
		}
	}
}

// validateGraphStructure checks the overall graph structure
func (v *WorkflowValidator) validateGraphStructure(nodes []WorkflowNode, connections []WorkflowConnection, nodeMap map[string]*WorkflowNode, result *WorkflowValidationResult) {
	if len(nodes) == 0 {
		return
	}

	// Build adjacency list and in-degree map
	edges := make(map[string][]string)
	inDegree := make(map[string]int)

	for _, node := range nodes {
		edges[node.ID] = []string{}
		inDegree[node.ID] = 0
	}

	for _, conn := range connections {
		if conn.SourceNodeID != "" && conn.TargetNodeID != "" {
			edges[conn.SourceNodeID] = append(edges[conn.SourceNodeID], conn.TargetNodeID)
			inDegree[conn.TargetNodeID]++
		}
	}

	// Find root nodes (no incoming edges)
	var rootNodes []string
	var triggerNodes []string

	for _, node := range nodes {
		if inDegree[node.ID] == 0 {
			rootNodes = append(rootNodes, node.ID)
			if isTriggerNode(node.Type) {
				triggerNodes = append(triggerNodes, node.ID)
			}
		}
	}

	// Validate at least one trigger node
	if len(triggerNodes) == 0 {
		result.AddError(WorkflowValidationError{
			Field:   "nodes",
			Code:    "NO_TRIGGER_NODE",
			Message: "Workflow must have at least one trigger node (trigger.manual, trigger.webhook, or trigger.schedule)",
		})
	}

	// Check for multiple trigger nodes (warning, not error)
	if len(triggerNodes) > 1 {
		result.AddError(WorkflowValidationError{
			Field:   "nodes",
			Code:    "MULTIPLE_TRIGGER_NODES",
			Message: fmt.Sprintf("Workflow has %d trigger nodes, only one will be used as entry point", len(triggerNodes)),
		})
	}

	// Check for orphan root nodes (non-trigger nodes with no incoming connections)
	for _, rootID := range rootNodes {
		node := nodeMap[rootID]
		if node != nil && !isTriggerNode(node.Type) {
			// Check if this node has outgoing connections
			if len(edges[rootID]) == 0 {
				result.AddError(WorkflowValidationError{
					Field:   "nodes",
					NodeID:  rootID,
					Code:    "ORPHAN_NODE",
					Message: fmt.Sprintf("Node '%s' is not connected to the workflow", node.Name),
				})
			}
		}
	}

	// Detect cycles using topological sort
	if hasCycle(nodes, edges, inDegree) {
		result.AddError(WorkflowValidationError{
			Field:   "connections",
			Code:    "CYCLE_DETECTED",
			Message: "Workflow contains a cycle which would cause infinite execution",
		})
	}

	// Check for unreachable nodes (nodes not reachable from any trigger)
	reachable := make(map[string]bool)
	for _, triggerID := range triggerNodes {
		v.markReachable(triggerID, edges, reachable)
	}

	for _, node := range nodes {
		if !reachable[node.ID] && !isTriggerNode(node.Type) {
			result.AddError(WorkflowValidationError{
				Field:   "nodes",
				NodeID:  node.ID,
				Code:    "UNREACHABLE_NODE",
				Message: fmt.Sprintf("Node '%s' is not reachable from any trigger", node.Name),
			})
		}
	}
}

// markReachable marks all nodes reachable from the given node
func (v *WorkflowValidator) markReachable(nodeID string, edges map[string][]string, reachable map[string]bool) {
	if reachable[nodeID] {
		return
	}
	reachable[nodeID] = true
	for _, targetID := range edges[nodeID] {
		v.markReachable(targetID, edges, reachable)
	}
}

// hasCycle detects if there's a cycle in the graph
func hasCycle(nodes []WorkflowNode, edges map[string][]string, originalInDegree map[string]int) bool {
	// Copy in-degree map
	inDegree := make(map[string]int)
	for k, v := range originalInDegree {
		inDegree[k] = v
	}

	// Find all nodes with no incoming edges
	var queue []string
	for _, node := range nodes {
		if inDegree[node.ID] == 0 {
			queue = append(queue, node.ID)
		}
	}

	visited := 0
	for len(queue) > 0 {
		nodeID := queue[0]
		queue = queue[1:]
		visited++

		for _, targetID := range edges[nodeID] {
			inDegree[targetID]--
			if inDegree[targetID] == 0 {
				queue = append(queue, targetID)
			}
		}
	}

	return visited != len(nodes)
}

// isTriggerNode checks if a node type is a trigger
func isTriggerNode(nodeType string) bool {
	return strings.HasPrefix(nodeType, "trigger.")
}

// ValidateWorkflow is a convenience function for quick validation
func ValidateWorkflow(nodes []WorkflowNode, connections []WorkflowConnection, typeChecker NodeTypeChecker) *WorkflowValidationResult {
	v := NewWorkflowValidator(typeChecker)
	return v.Validate(nodes, connections)
}

// ValidateWorkflowBasic validates without type checking (for quick validation)
func ValidateWorkflowBasic(nodes []WorkflowNode, connections []WorkflowConnection) *WorkflowValidationResult {
	v := NewWorkflowValidator(nil)
	return v.Validate(nodes, connections)
}
