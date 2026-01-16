package database

import (
	"context"
	"fmt"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type RedisNode struct{}

func NewRedisNode() *RedisNode {
	return &RedisNode{}
}

func (n *RedisNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	operation, _ := params["operation"].(string)

	switch operation {
	case "get":
		return n.get(ctx, params)
	case "set":
		return n.set(ctx, params)
	case "del":
		return n.del(ctx, params)
	case "hget":
		return n.hget(ctx, params)
	case "hset":
		return n.hset(ctx, params)
	default:
		return nil, fmt.Errorf("unsupported Redis operation: %s", operation)
	}
}

func (n *RedisNode) get(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	key, _ := params["key"].(string)
	return types.JSON{"key": key, "value": nil}, nil
}

func (n *RedisNode) set(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	key, _ := params["key"].(string)
	value := params["value"]
	return types.JSON{"key": key, "value": value, "success": true}, nil
}

func (n *RedisNode) del(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	keys, _ := params["keys"].([]interface{})
	return types.JSON{"keys": keys, "deleted": 0}, nil
}

func (n *RedisNode) hget(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	key, _ := params["key"].(string)
	field, _ := params["field"].(string)
	return types.JSON{"key": key, "field": field, "value": nil}, nil
}

func (n *RedisNode) hset(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	key, _ := params["key"].(string)
	field, _ := params["field"].(string)
	value := params["value"]
	return types.JSON{"key": key, "field": field, "value": value, "success": true}, nil
}

func (n *RedisNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "integration.redis",
		Name:        "Redis",
		Description: "Interact with Redis",
		Category:    "integration",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "object"}},
		Parameters:  []wtypes.NodeParameter{},
	}
}
