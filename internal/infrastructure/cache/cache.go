package cache

import (
	"context"
	"time"
)

// Cache defines the interface for caching operations
type Cache interface {
	// Get retrieves a value from cache
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores a value in cache with expiration
	Set(ctx context.Context, key string, value []byte, expiration time.Duration) error

	// Delete removes a value from cache
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists in cache
	Exists(ctx context.Context, key string) (bool, error)

	// GetMulti retrieves multiple values from cache
	GetMulti(ctx context.Context, keys []string) (map[string][]byte, error)

	// SetMulti stores multiple values in cache
	SetMulti(ctx context.Context, items map[string][]byte, expiration time.Duration) error

	// DeleteMulti removes multiple values from cache
	DeleteMulti(ctx context.Context, keys []string) error

	// Increment increments a numeric value
	Increment(ctx context.Context, key string, delta int64) (int64, error)

	// Decrement decrements a numeric value
	Decrement(ctx context.Context, key string, delta int64) (int64, error)

	// Flush clears all items from cache
	Flush(ctx context.Context) error

	// Close closes the cache connection
	Close() error
}

// ErrNotFound is returned when a key is not found in cache
var ErrNotFound = &NotFoundError{}

type NotFoundError struct{}

func (e *NotFoundError) Error() string {
	return "cache: key not found"
}
