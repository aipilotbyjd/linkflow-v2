package crm

import (
	"context"
	"fmt"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type HubSpotNode struct{}

func NewHubSpotNode() *HubSpotNode {
	return &HubSpotNode{}
}

func (n *HubSpotNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	operation, _ := params["operation"].(string)
	objectType, _ := params["object_type"].(string)

	switch operation {
	case "get":
		return n.getObject(ctx, objectType, params)
	case "create":
		return n.createObject(ctx, objectType, params)
	case "update":
		return n.updateObject(ctx, objectType, params)
	case "search":
		return n.searchObjects(ctx, objectType, params)
	default:
		return nil, fmt.Errorf("unsupported HubSpot operation: %s", operation)
	}
}

func (n *HubSpotNode) getObject(ctx context.Context, objectType string, params map[string]interface{}) (types.JSON, error) {
	objectID, _ := params["object_id"].(string)
	return types.JSON{"object_type": objectType, "object_id": objectID, "object": nil}, nil
}

func (n *HubSpotNode) createObject(ctx context.Context, objectType string, params map[string]interface{}) (types.JSON, error) {
	properties, _ := params["properties"].(map[string]interface{})
	return types.JSON{"object_type": objectType, "properties": properties, "id": "", "success": true}, nil
}

func (n *HubSpotNode) updateObject(ctx context.Context, objectType string, params map[string]interface{}) (types.JSON, error) {
	objectID, _ := params["object_id"].(string)
	properties, _ := params["properties"].(map[string]interface{})
	return types.JSON{"object_type": objectType, "object_id": objectID, "properties": properties, "success": true}, nil
}

func (n *HubSpotNode) searchObjects(ctx context.Context, objectType string, params map[string]interface{}) (types.JSON, error) {
	filters, _ := params["filters"].([]interface{})
	return types.JSON{"object_type": objectType, "filters": filters, "results": []interface{}{}, "total": 0}, nil
}

func (n *HubSpotNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "integration.hubspot",
		Name:        "HubSpot",
		Description: "Integrate with HubSpot CRM",
		Category:    "integration",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "object"}},
		Parameters:  []wtypes.NodeParameter{},
	}
}
