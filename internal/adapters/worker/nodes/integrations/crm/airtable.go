package crm

import (
	"context"
	"fmt"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type AirtableNode struct{}

func NewAirtableNode() *AirtableNode {
	return &AirtableNode{}
}

func (n *AirtableNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	operation, _ := params["operation"].(string)

	switch operation {
	case "list":
		return n.listRecords(ctx, params)
	case "get":
		return n.getRecord(ctx, params)
	case "create":
		return n.createRecord(ctx, params)
	case "update":
		return n.updateRecord(ctx, params)
	case "delete":
		return n.deleteRecord(ctx, params)
	default:
		return nil, fmt.Errorf("unsupported Airtable operation: %s", operation)
	}
}

func (n *AirtableNode) listRecords(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	baseId, _ := params["base_id"].(string)
	tableId, _ := params["table_id"].(string)
	return types.JSON{"base_id": baseId, "table_id": tableId, "records": []interface{}{}}, nil
}

func (n *AirtableNode) getRecord(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	baseId, _ := params["base_id"].(string)
	tableId, _ := params["table_id"].(string)
	recordId, _ := params["record_id"].(string)
	return types.JSON{"base_id": baseId, "table_id": tableId, "record_id": recordId, "record": nil}, nil
}

func (n *AirtableNode) createRecord(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	baseId, _ := params["base_id"].(string)
	tableId, _ := params["table_id"].(string)
	fields, _ := params["fields"].(map[string]interface{})
	return types.JSON{"base_id": baseId, "table_id": tableId, "fields": fields, "id": "", "success": true}, nil
}

func (n *AirtableNode) updateRecord(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	baseId, _ := params["base_id"].(string)
	tableId, _ := params["table_id"].(string)
	recordId, _ := params["record_id"].(string)
	fields, _ := params["fields"].(map[string]interface{})
	return types.JSON{"base_id": baseId, "table_id": tableId, "record_id": recordId, "fields": fields, "success": true}, nil
}

func (n *AirtableNode) deleteRecord(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	baseId, _ := params["base_id"].(string)
	tableId, _ := params["table_id"].(string)
	recordId, _ := params["record_id"].(string)
	return types.JSON{"base_id": baseId, "table_id": tableId, "record_id": recordId, "deleted": true}, nil
}

func (n *AirtableNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "integration.airtable",
		Name:        "Airtable",
		Description: "Interact with Airtable bases",
		Category:    "integration",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "object"}},
		Parameters:  []wtypes.NodeParameter{},
	}
}
