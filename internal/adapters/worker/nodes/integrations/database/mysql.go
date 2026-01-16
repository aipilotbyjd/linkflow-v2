package database

import (
	"context"
	"fmt"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type MySQLNode struct{}

func NewMySQLNode() *MySQLNode {
	return &MySQLNode{}
}

func (n *MySQLNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	operation, _ := params["operation"].(string)

	switch operation {
	case "query":
		return n.executeQuery(ctx, params)
	case "insert":
		return n.executeInsert(ctx, params)
	case "update":
		return n.executeUpdate(ctx, params)
	case "delete":
		return n.executeDelete(ctx, params)
	default:
		return nil, fmt.Errorf("unsupported MySQL operation: %s", operation)
	}
}

func (n *MySQLNode) executeQuery(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	query, _ := params["query"].(string)

	return types.JSON{
		"query":   query,
		"rows":    []interface{}{},
		"count":   0,
		"success": true,
	}, nil
}

func (n *MySQLNode) executeInsert(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	table, _ := params["table"].(string)
	data, _ := params["data"].(map[string]interface{})

	return types.JSON{
		"table":         table,
		"data":          data,
		"affected_rows": 1,
		"success":       true,
	}, nil
}

func (n *MySQLNode) executeUpdate(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	table, _ := params["table"].(string)
	data, _ := params["data"].(map[string]interface{})
	where, _ := params["where"].(string)

	return types.JSON{
		"table":         table,
		"data":          data,
		"where":         where,
		"affected_rows": 0,
		"success":       true,
	}, nil
}

func (n *MySQLNode) executeDelete(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	table, _ := params["table"].(string)
	where, _ := params["where"].(string)

	return types.JSON{
		"table":         table,
		"where":         where,
		"affected_rows": 0,
		"success":       true,
	}, nil
}

func (n *MySQLNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "integration.mysql",
		Name:        "MySQL",
		Description: "Query MySQL databases",
		Category:    "integration",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "object"}},
		Parameters:  []wtypes.NodeParameter{},
	}
}
