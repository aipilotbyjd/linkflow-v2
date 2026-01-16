package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrLockNotAcquired = errors.New("lock not acquired")
	ErrLockNotHeld     = errors.New("lock not held by this instance")
)

// Lock represents a distributed lock
type Lock struct {
	client   *Client
	key      string
	value    string
	ttl      time.Duration
	acquired bool
}

// LockManager manages distributed locks
type LockManager struct {
	client *Client
	prefix string
}

// NewLockManager creates a new lock manager
func NewLockManager(client *Client) *LockManager {
	return &LockManager{
		client: client,
		prefix: "lock:",
	}
}

// Acquire attempts to acquire a lock
func (m *LockManager) Acquire(ctx context.Context, key string, ttl time.Duration) (*Lock, error) {
	lockKey := m.prefix + key
	lockValue := uuid.New().String()

	acquired, err := m.client.client.SetNX(ctx, lockKey, lockValue, ttl).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}

	if !acquired {
		return nil, ErrLockNotAcquired
	}

	return &Lock{
		client:   m.client,
		key:      lockKey,
		value:    lockValue,
		ttl:      ttl,
		acquired: true,
	}, nil
}

// TryAcquire attempts to acquire a lock with retries
func (m *LockManager) TryAcquire(ctx context.Context, key string, ttl time.Duration, maxRetries int, retryDelay time.Duration) (*Lock, error) {
	for i := 0; i < maxRetries; i++ {
		lock, err := m.Acquire(ctx, key, ttl)
		if err == nil {
			return lock, nil
		}

		if !errors.Is(err, ErrLockNotAcquired) {
			return nil, err
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(retryDelay):
			continue
		}
	}

	return nil, ErrLockNotAcquired
}

// WithLock executes a function while holding a lock
func (m *LockManager) WithLock(ctx context.Context, key string, ttl time.Duration, fn func() error) error {
	lock, err := m.Acquire(ctx, key, ttl)
	if err != nil {
		return err
	}
	defer lock.Release(ctx)

	return fn()
}

// Release releases the lock
func (l *Lock) Release(ctx context.Context) error {
	if !l.acquired {
		return nil
	}

	// Use Lua script to ensure we only delete if we own the lock
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`

	result, err := l.client.client.Eval(ctx, script, []string{l.key}, l.value).Int64()
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	if result == 0 {
		return ErrLockNotHeld
	}

	l.acquired = false
	return nil
}

// Extend extends the lock TTL
func (l *Lock) Extend(ctx context.Context, ttl time.Duration) error {
	if !l.acquired {
		return ErrLockNotHeld
	}

	// Use Lua script to ensure we only extend if we own the lock
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("pexpire", KEYS[1], ARGV[2])
		else
			return 0
		end
	`

	result, err := l.client.client.Eval(ctx, script, []string{l.key}, l.value, ttl.Milliseconds()).Int64()
	if err != nil {
		return fmt.Errorf("failed to extend lock: %w", err)
	}

	if result == 0 {
		l.acquired = false
		return ErrLockNotHeld
	}

	l.ttl = ttl
	return nil
}

// IsHeld checks if the lock is still held
func (l *Lock) IsHeld(ctx context.Context) bool {
	if !l.acquired {
		return false
	}

	val, err := l.client.Get(ctx, l.key)
	if err != nil {
		return false
	}

	return val == l.value
}

// Semaphore provides a counting semaphore using Redis
type Semaphore struct {
	client   *Client
	key      string
	limit    int64
	acquired bool
	token    string
}

// NewSemaphore creates a new semaphore
func (m *LockManager) NewSemaphore(key string, limit int64) *Semaphore {
	return &Semaphore{
		client: m.client,
		key:    m.prefix + "sem:" + key,
		limit:  limit,
		token:  uuid.New().String(),
	}
}

// Acquire attempts to acquire a semaphore slot
func (s *Semaphore) Acquire(ctx context.Context, ttl time.Duration) error {
	now := time.Now().UnixNano()

	// Remove expired entries and try to add our token
	script := `
		-- Remove expired entries
		redis.call("zremrangebyscore", KEYS[1], "-inf", ARGV[1])
		
		-- Check current count
		local count = redis.call("zcard", KEYS[1])
		if count < tonumber(ARGV[2]) then
			-- Add our token with expiration time as score
			redis.call("zadd", KEYS[1], ARGV[3], ARGV[4])
			return 1
		else
			return 0
		end
	`

	expireAt := now + int64(ttl)
	result, err := s.client.client.Eval(ctx, script, []string{s.key}, now, s.limit, expireAt, s.token).Int64()
	if err != nil {
		return fmt.Errorf("failed to acquire semaphore: %w", err)
	}

	if result == 0 {
		return ErrLockNotAcquired
	}

	s.acquired = true
	return nil
}

// Release releases a semaphore slot
func (s *Semaphore) Release(ctx context.Context) error {
	if !s.acquired {
		return nil
	}

	err := s.client.client.ZRem(ctx, s.key, s.token).Err()
	if err != nil {
		return fmt.Errorf("failed to release semaphore: %w", err)
	}

	s.acquired = false
	return nil
}

// Available returns the number of available slots
func (s *Semaphore) Available(ctx context.Context) (int64, error) {
	// Remove expired entries first
	now := time.Now().UnixNano()
	s.client.client.ZRemRangeByScore(ctx, s.key, "-inf", fmt.Sprintf("%d", now))

	count, err := s.client.client.ZCard(ctx, s.key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get semaphore count: %w", err)
	}

	return s.limit - count, nil
}
