package nodes

import (
	"context"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// NodeInterface represents a workflow node that can be executed
type NodeInterface interface {
	// Execute runs the node logic and returns the output
	Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error)

	// Metadata returns information about the node type
	Metadata() wtypes.NodeMetadata
}

// Node is an alias for NodeInterface for backwards compatibility
type Node = NodeInterface

// Re-export types for convenience
type NodeMetadata = wtypes.NodeMetadata
type NodePort = wtypes.NodePort
type NodeParameter = wtypes.NodeParameter
type ParamOption = wtypes.ParamOption

// NodeFactory creates node instances
type NodeFactory func() Node

// NodeInput represents input data to a node
type NodeInput struct {
	Data       map[string]interface{}
	Parameters map[string]interface{}
	Credential map[string]interface{}
}

// GetString gets a string parameter value
func (i *NodeInput) GetString(key string) string {
	if v, ok := i.Parameters[key].(string); ok {
		return v
	}
	return ""
}

// GetInt gets an integer parameter value
func (i *NodeInput) GetInt(key string) int {
	switch v := i.Parameters[key].(type) {
	case int:
		return v
	case float64:
		return int(v)
	case int64:
		return int(v)
	}
	return 0
}

// GetBool gets a boolean parameter value
func (i *NodeInput) GetBool(key string) bool {
	if v, ok := i.Parameters[key].(bool); ok {
		return v
	}
	return false
}

// GetArray gets an array parameter value
func (i *NodeInput) GetArray(key string) []interface{} {
	if v, ok := i.Parameters[key].([]interface{}); ok {
		return v
	}
	return nil
}

// GetObject gets an object parameter value
func (i *NodeInput) GetObject(key string) map[string]interface{} {
	if v, ok := i.Parameters[key].(map[string]interface{}); ok {
		return v
	}
	return nil
}
