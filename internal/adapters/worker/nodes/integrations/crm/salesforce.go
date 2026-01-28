package crm

import (
	"context"
	"fmt"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type SalesforceNode struct{}

func NewSalesforceNode() *SalesforceNode {
	return &SalesforceNode{}
}

func (n *SalesforceNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	operation, _ := params["operation"].(string)

	switch operation {
	case "query":
		return n.query(ctx, params)
	case "get":
		return n.getRecord(ctx, params)
	case "create":
		return n.createRecord(ctx, params)
	case "update":
		return n.updateRecord(ctx, params)
	case "delete":
		return n.deleteRecord(ctx, params)
	default:
		return nil, fmt.Errorf("unsupported Salesforce operation: %s", operation)
	}
}

func (n *SalesforceNode) query(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	soql, _ := params["soql"].(string)
	return types.JSON{"soql": soql, "records": []interface{}{}, "totalSize": 0}, nil
}

func (n *SalesforceNode) getRecord(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	objectType, _ := params["object_type"].(string)
	recordID, _ := params["record_id"].(string)
	return types.JSON{"object_type": objectType, "record_id": recordID, "record": nil}, nil
}

func (n *SalesforceNode) createRecord(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	objectType, _ := params["object_type"].(string)
	data, _ := params["data"].(map[string]interface{})
	return types.JSON{"object_type": objectType, "data": data, "id": "", "success": true}, nil
}

func (n *SalesforceNode) updateRecord(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	objectType, _ := params["object_type"].(string)
	recordID, _ := params["record_id"].(string)
	data, _ := params["data"].(map[string]interface{})
	return types.JSON{"object_type": objectType, "record_id": recordID, "data": data, "success": true}, nil
}

func (n *SalesforceNode) deleteRecord(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	objectType, _ := params["object_type"].(string)
	recordID, _ := params["record_id"].(string)
	return types.JSON{"object_type": objectType, "record_id": recordID, "success": true}, nil
}

func (n *SalesforceNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "integration.salesforce",
		Name:        "Salesforce",
		Description: "Integrate with Salesforce CRM",
		Category:    "integration",
		Version:     "1.0.0",
		Icon:        "CloudUpload",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "object"}},
		Parameters:  []wtypes.NodeParameter{},
	}
}
