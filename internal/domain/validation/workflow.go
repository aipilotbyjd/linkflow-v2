package validation

import (
	"errors"
	"strings"
)

var (
	ErrWorkflowNameRequired    = errors.New("workflow name is required")
	ErrWorkflowNameTooLong     = errors.New("workflow name must be less than 255 characters")
	ErrWorkflowNoNodes         = errors.New("workflow must have at least one node")
	ErrWorkflowNoTrigger       = errors.New("workflow must have at least one trigger node")
	ErrWorkflowInvalidNode     = errors.New("workflow contains invalid node configuration")
	ErrWorkflowCircularDep     = errors.New("workflow contains circular dependencies")
	ErrWorkflowDisconnected    = errors.New("workflow has disconnected nodes")
)

// WorkflowInput represents input for workflow validation
type WorkflowInput struct {
	Name        string
	Description *string
	Nodes       []map[string]interface{}
	Connections []map[string]interface{}
}

// ValidateWorkflow validates workflow input
func ValidateWorkflow(input WorkflowInput) error {
	// Name validation
	if strings.TrimSpace(input.Name) == "" {
		return ErrWorkflowNameRequired
	}
	if len(input.Name) > 255 {
		return ErrWorkflowNameTooLong
	}

	// Nodes validation
	if len(input.Nodes) == 0 {
		return ErrWorkflowNoNodes
	}

	// Check for trigger node
	hasTrigger := false
	for _, node := range input.Nodes {
		nodeType, ok := node["type"].(string)
		if ok && strings.HasPrefix(nodeType, "trigger.") {
			hasTrigger = true
			break
		}
	}
	if !hasTrigger {
		return ErrWorkflowNoTrigger
	}

	return nil
}

// ValidateNode validates a single node configuration
func ValidateNode(node map[string]interface{}) error {
	// Check required fields
	if _, ok := node["id"]; !ok {
		return errors.New("node must have an id")
	}
	if _, ok := node["type"]; !ok {
		return errors.New("node must have a type")
	}
	if _, ok := node["name"]; !ok {
		return errors.New("node must have a name")
	}

	return nil
}
