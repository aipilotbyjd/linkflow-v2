package cache

import (
	"context"
	"time"
)

// NoopCache is a cache that does nothing
type NoopCache struct{}

func NewNoopCache() *NoopCache {
	return &NoopCache{}
}

func (c *NoopCache) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, ErrNotFound
}

func (c *NoopCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	return nil
}

func (c *NoopCache) Delete(ctx context.Context, key string) error {
	return nil
}

func (c *NoopCache) Exists(ctx context.Context, key string) (bool, error) {
	return false, nil
}

func (c *NoopCache) GetMulti(ctx context.Context, keys []string) (map[string][]byte, error) {
	return make(map[string][]byte), nil
}

func (c *NoopCache) SetMulti(ctx context.Context, items map[string][]byte, expiration time.Duration) error {
	return nil
}

func (c *NoopCache) DeleteMulti(ctx context.Context, keys []string) error {
	return nil
}

func (c *NoopCache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	return 0, nil
}

func (c *NoopCache) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	return 0, nil
}

func (c *NoopCache) Flush(ctx context.Context) error {
	return nil
}

func (c *NoopCache) Close() error {
	return nil
}
