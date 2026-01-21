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
		Description: "Execute queries and operations on PostgreSQL databases. Supports SELECT, INSERT, UPDATE, DELETE, and raw SQL.",
		Category:    "integration",
		Version:     "1.0.0",
		Icon:        "database",
		Color:       "#336791",
		Inputs: []wtypes.NodePort{
			{Name: "main", Type: "any", Description: "Input data for query parameters"},
		},
		Outputs: []wtypes.NodePort{
			{Name: "main", Type: "object", Description: "Query results or operation status"},
		},
		Parameters: []wtypes.NodeParameter{
			{
				Name:        "operation",
				DisplayName: "Operation",
				Type:        "options",
				Required:    true,
				Default:     "select",
				Description: "Database operation to perform",
				Options: []wtypes.ParamOption{
					{Name: "Select", Value: "select", Description: "Query data from table"},
					{Name: "Insert", Value: "insert", Description: "Insert new row(s)"},
					{Name: "Update", Value: "update", Description: "Update existing rows"},
					{Name: "Delete", Value: "delete", Description: "Delete rows"},
					{Name: "Execute Query", Value: "query", Description: "Run raw SQL query"},
					{Name: "Execute Function", Value: "function", Description: "Call stored function"},
				},
			},
			{
				Name:        "table",
				DisplayName: "Table",
				Type:        "string",
				Required:    false,
				Description: "Table name to operate on",
				Placeholder: "users",
				ShowIf:      "operation !== 'query' && operation !== 'function'",
			},
			{
				Name:        "columns",
				DisplayName: "Columns",
				Type:        "string",
				Required:    false,
				Default:     "*",
				Description: "Columns to select (comma-separated or *)",
				Placeholder: "id, name, email",
				ShowIf:      "operation === 'select'",
			},
			{
				Name:        "where",
				DisplayName: "WHERE Clause",
				Type:        "string",
				Required:    false,
				Description: "WHERE conditions (without WHERE keyword)",
				Placeholder: "status = 'active' AND created_at > $1",
				ShowIf:      "operation === 'select' || operation === 'update' || operation === 'delete'",
			},
			{
				Name:        "values",
				DisplayName: "Values",
				Type:        "json",
				Required:    false,
				Description: "Values to insert or update (object or array of objects)",
				Placeholder: `{"name": "John", "email": "john@example.com"}`,
				ShowIf:      "operation === 'insert' || operation === 'update'",
			},
			{
				Name:        "query",
				DisplayName: "SQL Query",
				Type:        "code",
				Required:    false,
				Description: "Raw SQL query to execute",
				Placeholder: "SELECT * FROM users WHERE id = $1",
				ShowIf:      "operation === 'query'",
			},
			{
				Name:        "params",
				DisplayName: "Query Parameters",
				Type:        "json",
				Required:    false,
				Description: "Parameters for parameterized queries ($1, $2, etc.)",
				Placeholder: `[1, "active"]`,
			},
			{
				Name:        "function_name",
				DisplayName: "Function Name",
				Type:        "string",
				Required:    false,
				Description: "Name of stored function to call",
				Placeholder: "calculate_total",
				ShowIf:      "operation === 'function'",
			},
			{
				Name:        "function_args",
				DisplayName: "Function Arguments",
				Type:        "json",
				Required:    false,
				Description: "Arguments for the stored function",
				ShowIf:      "operation === 'function'",
			},
			{
				Name:        "order_by",
				DisplayName: "Order By",
				Type:        "string",
				Required:    false,
				Description: "ORDER BY clause (without ORDER BY keyword)",
				Placeholder: "created_at DESC",
				ShowIf:      "operation === 'select'",
			},
			{
				Name:        "limit",
				DisplayName: "Limit",
				Type:        "number",
				Required:    false,
				Description: "Maximum number of rows to return",
				ShowIf:      "operation === 'select'",
			},
			{
				Name:        "offset",
				DisplayName: "Offset",
				Type:        "number",
				Required:    false,
				Description: "Number of rows to skip",
				ShowIf:      "operation === 'select'",
			},
			{
				Name:        "returning",
				DisplayName: "RETURNING Clause",
				Type:        "string",
				Required:    false,
				Description: "Columns to return after INSERT/UPDATE/DELETE",
				Placeholder: "id, created_at",
				ShowIf:      "operation === 'insert' || operation === 'update' || operation === 'delete'",
			},
			{
				Name:        "transaction",
				DisplayName: "Use Transaction",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Execute within a transaction (auto-rollback on error)",
			},
		},
		Credentials: []string{"postgres"},
	}
}
