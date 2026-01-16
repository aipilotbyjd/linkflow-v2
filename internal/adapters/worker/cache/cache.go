package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// WorkerCache provides caching for worker operations
type WorkerCache struct {
	client *redis.Client
	prefix string
}

// NewWorkerCache creates a new worker cache
func NewWorkerCache(client *redis.Client, prefix string) *WorkerCache {
	return &WorkerCache{
		client: client,
		prefix: prefix,
	}
}

// key generates a prefixed cache key
func (c *WorkerCache) key(parts ...string) string {
	key := c.prefix + ":"
	for i, part := range parts {
		if i > 0 {
			key += ":"
		}
		key += part
	}
	return key
}

// GetCredential retrieves cached credential data
func (c *WorkerCache) GetCredential(ctx context.Context, credentialID string) (map[string]interface{}, error) {
	key := c.key("credential", credentialID)

	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get credential from cache: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal cached credential: %w", err)
	}

	return result, nil
}

// SetCredential caches credential data
func (c *WorkerCache) SetCredential(ctx context.Context, credentialID string, data map[string]interface{}, ttl time.Duration) error {
	key := c.key("credential", credentialID)

	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal credential: %w", err)
	}

	if err := c.client.Set(ctx, key, bytes, ttl).Err(); err != nil {
		return fmt.Errorf("set credential in cache: %w", err)
	}

	return nil
}

// InvalidateCredential removes cached credential
func (c *WorkerCache) InvalidateCredential(ctx context.Context, credentialID string) error {
	key := c.key("credential", credentialID)
	return c.client.Del(ctx, key).Err()
}

// GetNodeResult retrieves cached node result
func (c *WorkerCache) GetNodeResult(ctx context.Context, executionID, nodeID string) (map[string]interface{}, error) {
	key := c.key("node", executionID, nodeID)

	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get node result from cache: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal cached node result: %w", err)
	}

	return result, nil
}

// SetNodeResult caches node execution result
func (c *WorkerCache) SetNodeResult(ctx context.Context, executionID, nodeID string, data map[string]interface{}, ttl time.Duration) error {
	key := c.key("node", executionID, nodeID)

	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal node result: %w", err)
	}

	if err := c.client.Set(ctx, key, bytes, ttl).Err(); err != nil {
		return fmt.Errorf("set node result in cache: %w", err)
	}

	return nil
}

// GetWorkflowDefinition retrieves cached workflow definition
func (c *WorkerCache) GetWorkflowDefinition(ctx context.Context, workflowID string, version int) (map[string]interface{}, error) {
	key := c.key("workflow", workflowID, fmt.Sprintf("v%d", version))

	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get workflow from cache: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal cached workflow: %w", err)
	}

	return result, nil
}

// SetWorkflowDefinition caches workflow definition
func (c *WorkerCache) SetWorkflowDefinition(ctx context.Context, workflowID string, version int, data map[string]interface{}, ttl time.Duration) error {
	key := c.key("workflow", workflowID, fmt.Sprintf("v%d", version))

	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal workflow: %w", err)
	}

	if err := c.client.Set(ctx, key, bytes, ttl).Err(); err != nil {
		return fmt.Errorf("set workflow in cache: %w", err)
	}

	return nil
}

// InvalidateWorkflow removes cached workflow definition
func (c *WorkerCache) InvalidateWorkflow(ctx context.Context, workflowID string) error {
	pattern := c.key("workflow", workflowID, "*")
	return c.deleteByPattern(ctx, pattern)
}

// deleteByPattern deletes all keys matching a pattern
func (c *WorkerCache) deleteByPattern(ctx context.Context, pattern string) error {
	iter := c.client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return fmt.Errorf("delete key %s: %w", iter.Val(), err)
		}
	}
	return iter.Err()
}
