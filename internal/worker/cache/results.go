package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// ResultCache caches node execution results
type ResultCache struct {
	redis *redis.Client
	ttl   time.Duration
}

// ResultCacheConfig configures the result cache
type ResultCacheConfig struct {
	TTL time.Duration
}

// NewResultCache creates a new result cache
func NewResultCache(redis *redis.Client, cfg ResultCacheConfig) *ResultCache {
	if cfg.TTL == 0 {
		cfg.TTL = 1 * time.Hour
	}

	return &ResultCache{
		redis: redis,
		ttl:   cfg.TTL,
	}
}

// Key generates a cache key for a node result
func (c *ResultCache) Key(executionID uuid.UUID, nodeID string, inputHash string) string {
	return fmt.Sprintf("node:result:%s:%s:%s", executionID, nodeID, inputHash)
}

// GlobalKey generates a workspace-level cache key (for idempotent nodes)
func (c *ResultCache) GlobalKey(workspaceID uuid.UUID, nodeType string, inputHash string) string {
	return fmt.Sprintf("node:global:%s:%s:%s", workspaceID, nodeType, inputHash)
}

// Get retrieves a cached result
func (c *ResultCache) Get(ctx context.Context, key string) (map[string]interface{}, bool) {
	data, err := c.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil, false
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, false
	}

	return result, true
}

// Set stores a result in cache
func (c *ResultCache) Set(ctx context.Context, key string, result map[string]interface{}) error {
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	return c.redis.Set(ctx, key, data, c.ttl).Err()
}

// SetWithTTL stores a result with custom TTL
func (c *ResultCache) SetWithTTL(ctx context.Context, key string, result map[string]interface{}, ttl time.Duration) error {
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	return c.redis.Set(ctx, key, data, ttl).Err()
}

// Delete removes a cached result
func (c *ResultCache) Delete(ctx context.Context, key string) error {
	return c.redis.Del(ctx, key).Err()
}

// DeleteByPattern deletes all keys matching a pattern
func (c *ResultCache) DeleteByPattern(ctx context.Context, pattern string) error {
	iter := c.redis.Scan(ctx, 0, pattern, 100).Iterator()
	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return err
	}

	if len(keys) > 0 {
		return c.redis.Del(ctx, keys...).Err()
	}

	return nil
}

// InvalidateExecution invalidates all cached results for an execution
func (c *ResultCache) InvalidateExecution(ctx context.Context, executionID uuid.UUID) error {
	pattern := fmt.Sprintf("node:result:%s:*", executionID)
	return c.DeleteByPattern(ctx, pattern)
}

// InvalidateWorkspace invalidates all cached results for a workspace
func (c *ResultCache) InvalidateWorkspace(ctx context.Context, workspaceID uuid.UUID) error {
	pattern := fmt.Sprintf("node:global:%s:*", workspaceID)
	return c.DeleteByPattern(ctx, pattern)
}

// HashInput creates a hash of input for cache key
func HashInput(input map[string]interface{}) string {
	data, _ := json.Marshal(input)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:8])
}

// CachedResult wraps a cached result with metadata
type CachedResult struct {
	Result    map[string]interface{} `json:"result"`
	CachedAt  time.Time              `json:"cached_at"`
	ExpiresAt time.Time              `json:"expires_at"`
	NodeType  string                 `json:"node_type"`
	InputHash string                 `json:"input_hash"`
}

// GetWithMetadata retrieves a cached result with metadata
func (c *ResultCache) GetWithMetadata(ctx context.Context, key string) (*CachedResult, bool) {
	data, err := c.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil, false
	}

	var cached CachedResult
	if err := json.Unmarshal(data, &cached); err != nil {
		// Try parsing as plain result for backward compatibility
		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, false
		}
		return &CachedResult{Result: result}, true
	}

	return &cached, true
}

// SetWithMetadata stores a result with metadata
func (c *ResultCache) SetWithMetadata(ctx context.Context, key string, result map[string]interface{}, nodeType, inputHash string) error {
	cached := CachedResult{
		Result:    result,
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(c.ttl),
		NodeType:  nodeType,
		InputHash: inputHash,
	}

	data, err := json.Marshal(cached)
	if err != nil {
		return err
	}

	return c.redis.Set(ctx, key, data, c.ttl).Err()
}

// ExecutionCache caches entire execution results
type ExecutionCache struct {
	redis *redis.Client
	ttl   time.Duration
}

// NewExecutionCache creates a new execution cache
func NewExecutionCache(redis *redis.Client, ttl time.Duration) *ExecutionCache {
	if ttl == 0 {
		ttl = 24 * time.Hour
	}
	return &ExecutionCache{redis: redis, ttl: ttl}
}

// CachedExecution represents a cached execution
type CachedExecution struct {
	ExecutionID uuid.UUID              `json:"execution_id"`
	WorkflowID  uuid.UUID              `json:"workflow_id"`
	Status      string                 `json:"status"`
	Output      map[string]interface{} `json:"output"`
	Duration    time.Duration          `json:"duration"`
	CachedAt    time.Time              `json:"cached_at"`
}

// Get retrieves a cached execution
func (c *ExecutionCache) Get(ctx context.Context, executionID uuid.UUID) (*CachedExecution, bool) {
	key := fmt.Sprintf("execution:cache:%s", executionID)
	data, err := c.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil, false
	}

	var cached CachedExecution
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, false
	}

	return &cached, true
}

// Set stores an execution in cache
func (c *ExecutionCache) Set(ctx context.Context, exec *CachedExecution) error {
	key := fmt.Sprintf("execution:cache:%s", exec.ExecutionID)
	exec.CachedAt = time.Now()

	data, err := json.Marshal(exec)
	if err != nil {
		return err
	}

	return c.redis.Set(ctx, key, data, c.ttl).Err()
}

// IsCacheable determines if a node type should be cached
func IsCacheable(nodeType string) bool {
	nonCacheable := map[string]bool{
		"action.http":             true,
		"action.sub_workflow":     true,
		"action.execute_workflow": true,
		"integration.slack":       true,
		"integration.email":       true,
		"integration.discord":     true,
		"integration.telegram":    true,
		"integration.github":      true,
		"integration.notion":      true,
		"integration.airtable":    true,
		"logic.wait":              true,
		"trigger.webhook":         true,
		"trigger.schedule":        true,
	}
	return !nonCacheable[nodeType]
}

// CacheStats holds cache statistics
type CacheStats struct {
	Hits       int64
	Misses     int64
	Size       int64
	HitRate    float64
	LastReset  time.Time
}

// CacheStatsCollector collects cache statistics
type CacheStatsCollector struct {
	redis *redis.Client
	key   string
}

// NewCacheStatsCollector creates a new stats collector
func NewCacheStatsCollector(redis *redis.Client, prefix string) *CacheStatsCollector {
	return &CacheStatsCollector{
		redis: redis,
		key:   fmt.Sprintf("cache:stats:%s", prefix),
	}
}

// RecordHit records a cache hit
func (c *CacheStatsCollector) RecordHit(ctx context.Context) {
	c.redis.HIncrBy(ctx, c.key, "hits", 1)
}

// RecordMiss records a cache miss
func (c *CacheStatsCollector) RecordMiss(ctx context.Context) {
	c.redis.HIncrBy(ctx, c.key, "misses", 1)
}

// GetStats retrieves cache statistics
func (c *CacheStatsCollector) GetStats(ctx context.Context) (*CacheStats, error) {
	data, err := c.redis.HGetAll(ctx, c.key).Result()
	if err != nil {
		return nil, err
	}

	stats := &CacheStats{}
	if v, ok := data["hits"]; ok {
		_, _ = fmt.Sscanf(v, "%d", &stats.Hits)
	}
	if v, ok := data["misses"]; ok {
		_, _ = fmt.Sscanf(v, "%d", &stats.Misses)
	}

	total := stats.Hits + stats.Misses
	if total > 0 {
		stats.HitRate = float64(stats.Hits) / float64(total)
	}

	return stats, nil
}

// Reset resets cache statistics
func (c *CacheStatsCollector) Reset(ctx context.Context) error {
	return c.redis.Del(ctx, c.key).Err()
}
