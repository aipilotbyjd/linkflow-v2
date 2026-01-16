package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
	prefix string
}

func NewRedisCache(client *redis.Client, prefix string) *RedisCache {
	return &RedisCache{
		client: client,
		prefix: prefix,
	}
}

func (c *RedisCache) prefixKey(key string) string {
	if c.prefix == "" {
		return key
	}
	return c.prefix + ":" + key
}

func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := c.client.Get(ctx, c.prefixKey(key)).Bytes()
	if err == redis.Nil {
		return nil, ErrNotFound
	}
	return val, err
}

func (c *RedisCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	return c.client.Set(ctx, c.prefixKey(key), value, expiration).Err()
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, c.prefixKey(key)).Err()
}

func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.client.Exists(ctx, c.prefixKey(key)).Result()
	return n > 0, err
}

func (c *RedisCache) GetMulti(ctx context.Context, keys []string) (map[string][]byte, error) {
	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = c.prefixKey(key)
	}

	vals, err := c.client.MGet(ctx, prefixedKeys...).Result()
	if err != nil {
		return nil, err
	}

	result := make(map[string][]byte)
	for i, val := range vals {
		if val != nil {
			if str, ok := val.(string); ok {
				result[keys[i]] = []byte(str)
			}
		}
	}
	return result, nil
}

func (c *RedisCache) SetMulti(ctx context.Context, items map[string][]byte, expiration time.Duration) error {
	pipe := c.client.Pipeline()
	for key, value := range items {
		pipe.Set(ctx, c.prefixKey(key), value, expiration)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (c *RedisCache) DeleteMulti(ctx context.Context, keys []string) error {
	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = c.prefixKey(key)
	}
	return c.client.Del(ctx, prefixedKeys...).Err()
}

func (c *RedisCache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	return c.client.IncrBy(ctx, c.prefixKey(key), delta).Result()
}

func (c *RedisCache) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	return c.client.DecrBy(ctx, c.prefixKey(key), delta).Result()
}

func (c *RedisCache) Flush(ctx context.Context) error {
	if c.prefix != "" {
		keys, err := c.client.Keys(ctx, c.prefix+":*").Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			return c.client.Del(ctx, keys...).Err()
		}
		return nil
	}
	return c.client.FlushDB(ctx).Err()
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}
