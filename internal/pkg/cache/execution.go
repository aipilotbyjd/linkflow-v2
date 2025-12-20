package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/redis/go-redis/v9"
)

type ExecutionCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewExecutionCache(client *redis.Client, ttl time.Duration) *ExecutionCache {
	if ttl == 0 {
		ttl = 24 * time.Hour
	}
	return &ExecutionCache{
		client: client,
		ttl:    ttl,
	}
}

type CachedResult struct {
	Output    models.JSON `json:"output"`
	Status    string      `json:"status"`
	CachedAt  time.Time   `json:"cached_at"`
	ExpiresAt time.Time   `json:"expires_at"`
}

func (c *ExecutionCache) generateKey(workflowID uuid.UUID, version int, inputHash string) string {
	return fmt.Sprintf("execution:cache:%s:%d:%s", workflowID.String(), version, inputHash)
}

func (c *ExecutionCache) hashInput(input models.JSON) string {
	data, _ := json.Marshal(input)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (c *ExecutionCache) Get(ctx context.Context, workflowID uuid.UUID, version int, input models.JSON) (*CachedResult, error) {
	inputHash := c.hashInput(input)
	key := c.generateKey(workflowID, version, inputHash)

	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var result CachedResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	if time.Now().After(result.ExpiresAt) {
		c.client.Del(ctx, key)
		return nil, nil
	}

	return &result, nil
}

func (c *ExecutionCache) Set(ctx context.Context, workflowID uuid.UUID, version int, input, output models.JSON) error {
	inputHash := c.hashInput(input)
	key := c.generateKey(workflowID, version, inputHash)

	now := time.Now()
	result := CachedResult{
		Output:    output,
		Status:    "cached",
		CachedAt:  now,
		ExpiresAt: now.Add(c.ttl),
	}

	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, c.ttl).Err()
}

func (c *ExecutionCache) Invalidate(ctx context.Context, workflowID uuid.UUID) error {
	pattern := fmt.Sprintf("execution:cache:%s:*", workflowID.String())
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}

	return iter.Err()
}

func (c *ExecutionCache) InvalidateVersion(ctx context.Context, workflowID uuid.UUID, version int) error {
	pattern := fmt.Sprintf("execution:cache:%s:%d:*", workflowID.String(), version)
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}

	return iter.Err()
}

// CredentialCache caches decrypted credentials
type CredentialCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewCredentialCache(client *redis.Client, ttl time.Duration) *CredentialCache {
	if ttl == 0 {
		ttl = 5 * time.Minute
	}
	return &CredentialCache{
		client: client,
		ttl:    ttl,
	}
}

func (c *CredentialCache) generateKey(credentialID uuid.UUID) string {
	return fmt.Sprintf("credential:cache:%s", credentialID.String())
}

func (c *CredentialCache) Get(ctx context.Context, credentialID uuid.UUID) (map[string]interface{}, error) {
	key := c.generateKey(credentialID)

	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *CredentialCache) Set(ctx context.Context, credentialID uuid.UUID, data map[string]interface{}) error {
	key := c.generateKey(credentialID)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, jsonData, c.ttl).Err()
}

func (c *CredentialCache) Invalidate(ctx context.Context, credentialID uuid.UUID) error {
	key := c.generateKey(credentialID)
	return c.client.Del(ctx, key).Err()
}

// NodeResultCache caches individual node results during execution
type NodeResultCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewNodeResultCache(client *redis.Client, ttl time.Duration) *NodeResultCache {
	if ttl == 0 {
		ttl = 1 * time.Hour
	}
	return &NodeResultCache{
		client: client,
		ttl:    ttl,
	}
}

func (c *NodeResultCache) generateKey(executionID uuid.UUID, nodeID string) string {
	return fmt.Sprintf("node:result:%s:%s", executionID.String(), nodeID)
}

func (c *NodeResultCache) Get(ctx context.Context, executionID uuid.UUID, nodeID string) (map[string]interface{}, error) {
	key := c.generateKey(executionID, nodeID)

	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *NodeResultCache) Set(ctx context.Context, executionID uuid.UUID, nodeID string, result map[string]interface{}) error {
	key := c.generateKey(executionID, nodeID)

	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, c.ttl).Err()
}

func (c *NodeResultCache) Clear(ctx context.Context, executionID uuid.UUID) error {
	pattern := fmt.Sprintf("node:result:%s:*", executionID.String())
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}

	return iter.Err()
}
