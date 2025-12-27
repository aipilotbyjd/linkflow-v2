package validator

import (
	"encoding/json"

	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/worker/core"
)

// ParseAndValidateWorkflow parses JSON arrays and validates the workflow
func ParseAndValidateWorkflow(nodesJSON, connectionsJSON models.JSONArray) (*WorkflowValidationResult, error) {
	nodes, connections, err := parseWorkflowData(nodesJSON, connectionsJSON)
	if err != nil {
		return nil, err
	}

	// Use the registered node type checker
	typeChecker := func(nodeType string) bool {
		return core.Get(nodeType) != nil
	}

	result := ValidateWorkflow(nodes, connections, typeChecker)

	// Also validate node parameters
	paramErrors := ValidateNodeParameters(nodes)
	for _, e := range paramErrors {
		result.AddError(WorkflowValidationError{
			Field:   "parameters." + e.Param,
			NodeID:  e.NodeID,
			Code:    e.Code,
			Message: e.Message,
		})
	}

	return result, nil
}

// parseWorkflowData parses nodes and connections from JSON arrays
func parseWorkflowData(nodesJSON, connectionsJSON models.JSONArray) ([]WorkflowNode, []WorkflowConnection, error) {
	var nodes []WorkflowNode
	var connections []WorkflowConnection

	// Parse nodes
	if len(nodesJSON) > 0 {
		nodesBytes, err := json.Marshal(nodesJSON)
		if err != nil {
			return nil, nil, err
		}
		if err := json.Unmarshal(nodesBytes, &nodes); err != nil {
			return nil, nil, err
		}
	}

	// Parse connections
	if len(connectionsJSON) > 0 {
		connBytes, err := json.Marshal(connectionsJSON)
		if err != nil {
			return nil, nil, err
		}
		if err := json.Unmarshal(connBytes, &connections); err != nil {
			return nil, nil, err
		}
	}

	return nodes, connections, nil
}

// ParseAndValidateWorkflowStrict performs strict validation with type checking
func ParseAndValidateWorkflowStrict(nodesJSON, connectionsJSON models.JSONArray) (*WorkflowValidationResult, error) {
	return ParseAndValidateWorkflow(nodesJSON, connectionsJSON)
}

// ParseAndValidateWorkflowBasic performs basic structure validation without type checking
func ParseAndValidateWorkflowBasic(nodesJSON, connectionsJSON models.JSONArray) (*WorkflowValidationResult, error) {
	var nodes []WorkflowNode
	var connections []WorkflowConnection

	// Parse nodes
	if len(nodesJSON) > 0 {
		nodesBytes, err := json.Marshal(nodesJSON)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(nodesBytes, &nodes); err != nil {
			return nil, err
		}
	}

	// Parse connections
	if len(connectionsJSON) > 0 {
		connBytes, err := json.Marshal(connectionsJSON)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(connBytes, &connections); err != nil {
			return nil, err
		}
	}

	return ValidateWorkflowBasic(nodes, connections), nil
}

// WorkflowValidationErrors converts validation result to DTO-compatible format
func WorkflowValidationErrors(result *WorkflowValidationResult) []ValidationError {
	if result.Valid {
		return nil
	}

	errors := make([]ValidationError, len(result.Errors))
	for i, e := range result.Errors {
		field := e.Field
		if e.NodeID != "" {
			field = "node:" + e.NodeID + "." + e.Field
		}
		errors[i] = ValidationError{
			Field:   field,
			Message: e.Message,
		}
	}
	return errors
}
