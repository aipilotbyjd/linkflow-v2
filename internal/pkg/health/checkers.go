package health

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// DatabaseChecker checks database connectivity
type DatabaseChecker struct {
	db   *gorm.DB
	name string
}

// NewDatabaseChecker creates a new database health checker
func NewDatabaseChecker(db *gorm.DB) *DatabaseChecker {
	return &DatabaseChecker{
		db:   db,
		name: "database",
	}
}

func (c *DatabaseChecker) Name() string {
	return c.name
}

func (c *DatabaseChecker) Check(ctx context.Context) error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	return nil
}

// RedisChecker checks Redis connectivity
type RedisChecker struct {
	client *redis.Client
	name   string
}

// NewRedisChecker creates a new Redis health checker
func NewRedisChecker(client *redis.Client) *RedisChecker {
	return &RedisChecker{
		client: client,
		name:   "redis",
	}
}

func (c *RedisChecker) Name() string {
	return c.name
}

func (c *RedisChecker) Check(ctx context.Context) error {
	if err := c.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}
	return nil
}

// CustomChecker allows creating custom health checks with a function
type CustomChecker struct {
	name    string
	checkFn func(ctx context.Context) error
}

// NewCustomChecker creates a new custom health checker
func NewCustomChecker(name string, checkFn func(ctx context.Context) error) *CustomChecker {
	return &CustomChecker{
		name:    name,
		checkFn: checkFn,
	}
}

func (c *CustomChecker) Name() string {
	return c.name
}

func (c *CustomChecker) Check(ctx context.Context) error {
	return c.checkFn(ctx)
}

// MemoryChecker checks memory usage
type MemoryChecker struct {
	maxMemoryMB int64
}

// NewMemoryChecker creates a memory usage checker
func NewMemoryChecker(maxMemoryMB int64) *MemoryChecker {
	return &MemoryChecker{maxMemoryMB: maxMemoryMB}
}

func (c *MemoryChecker) Name() string {
	return "memory"
}

func (c *MemoryChecker) Check(ctx context.Context) error {
	// Memory check can be implemented with runtime.MemStats
	// For now, always return healthy
	return nil
}
