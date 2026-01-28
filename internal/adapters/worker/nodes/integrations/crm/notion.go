package crm

import (
	"context"
	"fmt"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type NotionNode struct{}

func NewNotionNode() *NotionNode {
	return &NotionNode{}
}

func (n *NotionNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	operation, _ := params["operation"].(string)

	switch operation {
	case "get_page":
		return n.getPage(ctx, params)
	case "create_page":
		return n.createPage(ctx, params)
	case "query_database":
		return n.queryDatabase(ctx, params)
	case "search":
		return n.search(ctx, params)
	default:
		return nil, fmt.Errorf("unsupported Notion operation: %s", operation)
	}
}

func (n *NotionNode) getPage(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	pageID, _ := params["page_id"].(string)
	return types.JSON{"page_id": pageID, "page": nil}, nil
}

func (n *NotionNode) createPage(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	parentID, _ := params["parent_id"].(string)
	properties, _ := params["properties"].(map[string]interface{})
	return types.JSON{"parent_id": parentID, "properties": properties, "id": "", "url": ""}, nil
}

func (n *NotionNode) queryDatabase(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	databaseID, _ := params["database_id"].(string)
	filter, _ := params["filter"].(map[string]interface{})
	return types.JSON{"database_id": databaseID, "filter": filter, "results": []interface{}{}, "has_more": false}, nil
}

func (n *NotionNode) search(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	query, _ := params["query"].(string)
	return types.JSON{"query": query, "results": []interface{}{}, "has_more": false}, nil
}

func (n *NotionNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "integration.notion",
		Name:        "Notion",
		Description: "Interact with Notion pages",
		Category:    "integration",
		Version:     "1.0.0",
		Icon:        "Notion",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "object"}},
		Parameters:  []wtypes.NodeParameter{},
	}
}
