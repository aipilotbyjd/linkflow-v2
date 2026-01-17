package database

import (
	"context"
	"fmt"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type PostgresNode struct{}

func NewPostgresNode() *PostgresNode {
	return &PostgresNode{}
}

func (n *PostgresNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})

	query, _ := params["query"].(string)
	operation, _ := params["operation"].(string)

	if query == "" && operation == "" {
		return nil, fmt.Errorf("query or operation is required")
	}

	// PostgreSQL operations require credentials from the runtime
	// This returns a placeholder - full implementation requires database driver
	return types.JSON{
		"executed":  false,
		"operation": operation,
		"query":     query,
		"message":   "PostgreSQL requires database credential configuration",
	}, nil
}

func (n *PostgresNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "integration.postgres",
		Name:        "PostgreSQL",
		Description: "Query PostgreSQL databases",
		Category:    "integration",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "object"}},
		Parameters:  []wtypes.NodeParameter{},
	}
}
