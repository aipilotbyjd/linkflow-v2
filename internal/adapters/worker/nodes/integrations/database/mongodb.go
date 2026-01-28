package database

import (
	"context"
	"fmt"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type MongoDBNode struct{}

func NewMongoDBNode() *MongoDBNode {
	return &MongoDBNode{}
}

func (n *MongoDBNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	operation, _ := params["operation"].(string)

	switch operation {
	case "find":
		return n.find(ctx, params)
	case "find_one":
		return n.findOne(ctx, params)
	case "insert_one":
		return n.insertOne(ctx, params)
	case "update_one":
		return n.updateOne(ctx, params)
	case "delete_one":
		return n.deleteOne(ctx, params)
	default:
		return nil, fmt.Errorf("unsupported MongoDB operation: %s", operation)
	}
}

func (n *MongoDBNode) find(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	collection, _ := params["collection"].(string)
	filter, _ := params["filter"].(map[string]interface{})

	return types.JSON{
		"collection": collection,
		"filter":     filter,
		"documents":  []interface{}{},
		"count":      0,
	}, nil
}

func (n *MongoDBNode) findOne(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	collection, _ := params["collection"].(string)
	filter, _ := params["filter"].(map[string]interface{})

	return types.JSON{
		"collection": collection,
		"filter":     filter,
		"document":   nil,
		"found":      false,
	}, nil
}

func (n *MongoDBNode) insertOne(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	collection, _ := params["collection"].(string)
	document, _ := params["document"].(map[string]interface{})

	return types.JSON{
		"collection":  collection,
		"document":    document,
		"inserted_id": "",
		"success":     true,
	}, nil
}

func (n *MongoDBNode) updateOne(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	collection, _ := params["collection"].(string)
	filter, _ := params["filter"].(map[string]interface{})
	update, _ := params["update"].(map[string]interface{})

	return types.JSON{
		"collection":     collection,
		"filter":         filter,
		"update":         update,
		"modified_count": 0,
		"success":        true,
	}, nil
}

func (n *MongoDBNode) deleteOne(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	collection, _ := params["collection"].(string)
	filter, _ := params["filter"].(map[string]interface{})

	return types.JSON{
		"collection":    collection,
		"filter":        filter,
		"deleted_count": 0,
		"success":       true,
	}, nil
}

func (n *MongoDBNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "integration.mongodb",
		Name:        "MongoDB",
		Description: "Query MongoDB databases",
		Category:    "integration",
		Version:     "1.0.0",
		Icon:        "Database01",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "object"}},
		Parameters:  []wtypes.NodeParameter{},
	}
}
