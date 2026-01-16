package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// CacheStore provides caching functionality using Redis
type CacheStore struct {
	client *Client
	prefix string
}

// NewCacheStore creates a new cache store
func NewCacheStore(client *Client, prefix string) *CacheStore {
	if prefix == "" {
		prefix = "cache:"
	}
	return &CacheStore{
		client: client,
		prefix: prefix,
	}
}

// Get retrieves a value from cache
func (c *CacheStore) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, c.prefix+key)
}

// GetJSON retrieves and unmarshals a JSON value from cache
func (c *CacheStore) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := c.Get(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), dest)
}

// Set stores a value in cache with TTL
func (c *CacheStore) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.client.Set(ctx, c.prefix+key, value, ttl)
}

// SetJSON marshals and stores a JSON value in cache
func (c *CacheStore) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	return c.Set(ctx, key, string(data), ttl)
}

// Delete removes a value from cache
func (c *CacheStore) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, c.prefix+key)
}

// DeletePattern removes all keys matching a pattern
func (c *CacheStore) DeletePattern(ctx context.Context, pattern string) error {
	iter := c.client.client.Scan(ctx, 0, c.prefix+pattern, 100).Iterator()
	for iter.Next(ctx) {
		if err := c.client.client.Del(ctx, iter.Val()).Err(); err != nil {
			return fmt.Errorf("failed to delete key %s: %w", iter.Val(), err)
		}
	}
	return iter.Err()
}

// Exists checks if a key exists in cache
func (c *CacheStore) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.client.Exists(ctx, c.prefix+key)
	return count > 0, err
}

// GetOrSet gets a value from cache or sets it using the provided function
func (c *CacheStore) GetOrSet(ctx context.Context, key string, ttl time.Duration, fn func() (interface{}, error)) (string, error) {
	// Try to get from cache
	val, err := c.Get(ctx, key)
	if err == nil {
		return val, nil
	}

	// Call function to get value
	result, err := fn()
	if err != nil {
		return "", err
	}

	// Store in cache
	var strVal string
	switch v := result.(type) {
	case string:
		strVal = v
	case []byte:
		strVal = string(v)
	default:
		data, err := json.Marshal(result)
		if err != nil {
			return "", fmt.Errorf("failed to marshal result: %w", err)
		}
		strVal = string(data)
	}

	if err := c.Set(ctx, key, strVal, ttl); err != nil {
		// Log but don't fail - we have the value
		return strVal, nil
	}

	return strVal, nil
}

// Increment increments a counter
func (c *CacheStore) Increment(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, c.prefix+key)
}

// IncrementWithTTL increments a counter and sets TTL if it's a new key
func (c *CacheStore) IncrementWithTTL(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	val, err := c.client.Incr(ctx, c.prefix+key)
	if err != nil {
		return 0, err
	}

	// Set TTL if this is the first increment
	if val == 1 {
		c.client.Expire(ctx, c.prefix+key, ttl)
	}

	return val, nil
}

// SetNX sets a value only if it doesn't exist (for locking)
func (c *CacheStore) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	return c.client.client.SetNX(ctx, c.prefix+key, value, ttl).Result()
}

// GetMulti retrieves multiple values from cache
func (c *CacheStore) GetMulti(ctx context.Context, keys ...string) (map[string]string, error) {
	if len(keys) == 0 {
		return make(map[string]string), nil
	}

	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = c.prefix + key
	}

	values, err := c.client.client.MGet(ctx, prefixedKeys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get multiple keys: %w", err)
	}

	result := make(map[string]string)
	for i, val := range values {
		if val != nil {
			result[keys[i]] = val.(string)
		}
	}

	return result, nil
}

// SetMulti stores multiple values in cache
func (c *CacheStore) SetMulti(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	pipe := c.client.client.Pipeline()

	for key, value := range items {
		pipe.Set(ctx, c.prefix+key, value, ttl)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to set multiple keys: %w", err)
	}

	return nil
}

// TTL returns the remaining TTL for a key
func (c *CacheStore) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.client.TTL(ctx, c.prefix+key).Result()
}

// Touch updates the TTL of a key without changing its value
func (c *CacheStore) Touch(ctx context.Context, key string, ttl time.Duration) error {
	return c.client.Expire(ctx, c.prefix+key, ttl)
}
