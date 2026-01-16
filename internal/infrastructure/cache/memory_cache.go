package cache

import (
	"context"
	"sync"
	"time"
)

type cacheItem struct {
	value      []byte
	expiration time.Time
}

func (i cacheItem) isExpired() bool {
	if i.expiration.IsZero() {
		return false
	}
	return time.Now().After(i.expiration)
}

type MemoryCache struct {
	items map[string]cacheItem
	mu    sync.RWMutex
}

func NewMemoryCache() *MemoryCache {
	c := &MemoryCache{
		items: make(map[string]cacheItem),
	}
	go c.cleanup()
	return c
}

func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		c.mu.Lock()
		for key, item := range c.items {
			if item.isExpired() {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

func (c *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.items[key]
	if !ok || item.isExpired() {
		return nil, ErrNotFound
	}
	return item.value, nil
}

func (c *MemoryCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var exp time.Time
	if expiration > 0 {
		exp = time.Now().Add(expiration)
	}

	c.items[key] = cacheItem{
		value:      value,
		expiration: exp,
	}
	return nil
}

func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
	return nil
}

func (c *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.items[key]
	if !ok || item.isExpired() {
		return false, nil
	}
	return true, nil
}

func (c *MemoryCache) GetMulti(ctx context.Context, keys []string) (map[string][]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string][]byte)
	for _, key := range keys {
		if item, ok := c.items[key]; ok && !item.isExpired() {
			result[key] = item.value
		}
	}
	return result, nil
}

func (c *MemoryCache) SetMulti(ctx context.Context, items map[string][]byte, expiration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var exp time.Time
	if expiration > 0 {
		exp = time.Now().Add(expiration)
	}

	for key, value := range items {
		c.items[key] = cacheItem{
			value:      value,
			expiration: exp,
		}
	}
	return nil
}

func (c *MemoryCache) DeleteMulti(ctx context.Context, keys []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, key := range keys {
		delete(c.items, key)
	}
	return nil
}

func (c *MemoryCache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	// Memory cache doesn't support atomic increment
	return 0, nil
}

func (c *MemoryCache) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	// Memory cache doesn't support atomic decrement
	return 0, nil
}

func (c *MemoryCache) Flush(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]cacheItem)
	return nil
}

func (c *MemoryCache) Close() error {
	return nil
}
