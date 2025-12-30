package config

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// =============================================================================
// Redis Options Builder
// =============================================================================

// RedisOptions returns go-redis client options
func (c *RedisConfig) RedisOptions() *redis.Options {
	opts := &redis.Options{
		Addr:            c.Addr(),
		Password:        c.Password,
		DB:              c.DB,
		PoolSize:        c.PoolSize,
		MinIdleConns:    c.MinIdleConns,
		MaxRetries:      c.MaxRetries,
		DialTimeout:     c.DialTimeout,
		ReadTimeout:     c.ReadTimeout,
		WriteTimeout:    c.WriteTimeout,
		PoolTimeout:     c.PoolTimeout,
		ConnMaxLifetime: c.MaxConnAge,
	}

	if c.TLS {
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	return opts
}

// RedisClusterOptions returns go-redis cluster options
func (c *RedisConfig) RedisClusterOptions() *redis.ClusterOptions {
	opts := &redis.ClusterOptions{
		Addrs:           c.ClusterAddrs,
		Password:        c.Password,
		PoolSize:        c.PoolSize,
		MinIdleConns:    c.MinIdleConns,
		MaxRetries:      c.MaxRetries,
		DialTimeout:     c.DialTimeout,
		ReadTimeout:     c.ReadTimeout,
		WriteTimeout:    c.WriteTimeout,
		PoolTimeout:     c.PoolTimeout,
		ConnMaxLifetime: c.MaxConnAge,
	}

	if c.TLS {
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	return opts
}

// RedisFailoverOptions returns go-redis sentinel options
func (c *RedisConfig) RedisFailoverOptions() *redis.FailoverOptions {
	opts := &redis.FailoverOptions{
		MasterName:      c.SentinelMaster,
		SentinelAddrs:   c.SentinelAddrs,
		Password:        c.Password,
		DB:              c.DB,
		PoolSize:        c.PoolSize,
		MinIdleConns:    c.MinIdleConns,
		MaxRetries:      c.MaxRetries,
		DialTimeout:     c.DialTimeout,
		ReadTimeout:     c.ReadTimeout,
		WriteTimeout:    c.WriteTimeout,
		PoolTimeout:     c.PoolTimeout,
		ConnMaxLifetime: c.MaxConnAge,
	}

	if c.TLS {
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	return opts
}

// =============================================================================
// Asynq Options Builder
// =============================================================================

// AsynqRedisClientOpt returns asynq Redis client options
func (c *RedisConfig) AsynqRedisClientOpt() asynq.RedisClientOpt {
	opt := asynq.RedisClientOpt{
		Addr:     c.Addr(),
		Password: c.Password,
		DB:       c.DB,
	}

	if c.TLS {
		opt.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	return opt
}

// AsynqRedisClusterClientOpt returns asynq Redis cluster options
func (c *RedisConfig) AsynqRedisClusterClientOpt() asynq.RedisClusterClientOpt {
	opt := asynq.RedisClusterClientOpt{
		Addrs:    c.ClusterAddrs,
		Password: c.Password,
	}

	if c.TLS {
		opt.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	return opt
}

// AsynqRedisFailoverClientOpt returns asynq Redis sentinel options
func (c *RedisConfig) AsynqRedisFailoverClientOpt() asynq.RedisFailoverClientOpt {
	opt := asynq.RedisFailoverClientOpt{
		MasterName:    c.SentinelMaster,
		SentinelAddrs: c.SentinelAddrs,
		Password:      c.Password,
		DB:            c.DB,
	}

	if c.TLS {
		opt.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	return opt
}

// AsynqServerConfig returns asynq server configuration
func (c *QueueConfig) AsynqServerConfig() asynq.Config {
	return asynq.Config{
		Concurrency:        c.Concurrency,
		StrictPriority:     c.StrictPriority,
		ShutdownTimeout:    c.ShutdownTimeout,
		HealthCheckInterval: c.HealthCheckInterval,
		Queues: map[string]int{
			"critical": c.Queues.Critical,
			"default":  c.Queues.Default,
			"low":      c.Queues.Low,
		},
	}
}

// =============================================================================
// Database Options Builder
// =============================================================================

// GormConfig returns GORM configuration
func (c *DatabaseConfig) GormConfig(debug bool) *gorm.Config {
	logLevel := gormlogger.Silent
	if debug || c.LogQueries {
		logLevel = gormlogger.Info
	}

	return &gorm.Config{
		Logger: gormlogger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		PrepareStmt:                              true,
		DisableForeignKeyConstraintWhenMigrating: false,
	}
}

// ConnectionPoolConfig returns connection pool settings as a map
func (c *DatabaseConfig) ConnectionPoolConfig() map[string]interface{} {
	return map[string]interface{}{
		"max_open_conns":      c.MaxOpenConns,
		"max_idle_conns":      c.MaxIdleConns,
		"conn_max_lifetime":   c.ConnMaxLifetime,
		"conn_max_idle_time":  c.ConnMaxIdleTime,
	}
}

// =============================================================================
// Server Options
// =============================================================================

// HTTPServerConfig returns HTTP server configuration
func (c *ServerConfig) HTTPServerConfig() map[string]interface{} {
	return map[string]interface{}{
		"addr":            fmt.Sprintf("%s:%d", c.Host, c.Port),
		"read_timeout":    c.ReadTimeout,
		"write_timeout":   c.WriteTimeout,
		"idle_timeout":    c.IdleTimeout,
		"max_header_bytes": c.MaxHeaderBytes,
	}
}

// =============================================================================
// CORS Options
// =============================================================================

// CORSOptions returns CORS middleware options
func (c *CORSConfig) CORSOptions() map[string]interface{} {
	return map[string]interface{}{
		"allowed_origins":   c.AllowedOrigins,
		"allowed_methods":   c.AllowedMethods,
		"allowed_headers":   c.AllowedHeaders,
		"exposed_headers":   c.ExposedHeaders,
		"allow_credentials": c.AllowCredentials,
		"max_age":           c.MaxAge,
	}
}

// =============================================================================
// S3 Options
// =============================================================================

// S3Config returns S3 configuration for AWS SDK
func (c *S3Config) S3Options() map[string]interface{} {
	return map[string]interface{}{
		"endpoint":          c.Endpoint,
		"region":            c.Region,
		"bucket":            c.Bucket,
		"use_ssl":           c.UseSSL,
		"path_style":        c.PathStyle,
		"presigned_expiry":  c.PresignedExpiry,
		"upload_part_size":  c.UploadPartSize,
		"max_upload_parts":  c.MaxUploadParts,
	}
}

// IsAWS returns true if using AWS S3 (no custom endpoint)
func (c *S3Config) IsAWS() bool {
	return c.Endpoint == ""
}

// IsMinIO returns true if likely using MinIO
func (c *S3Config) IsMinIO() bool {
	return c.Endpoint != "" && c.PathStyle
}

// =============================================================================
// Rate Limit Options
// =============================================================================

// RateLimitOptions returns rate limit configuration
func (c *RateLimitConfig) RateLimitOptions() map[string]interface{} {
	return map[string]interface{}{
		"enabled":             c.Enabled,
		"requests_per_second": c.RequestsPerSecond,
		"burst":               c.Burst,
		"by_ip":               c.ByIP,
		"by_user":             c.ByUser,
		"exclude_paths":       c.ExcludePaths,
		"custom_limits":       c.CustomLimits,
	}
}

// =============================================================================
// Tracing Options
// =============================================================================

// TracingOptions returns tracing configuration
func (c *TracingConfig) TracingOptions() map[string]interface{} {
	return map[string]interface{}{
		"enabled":      c.Enabled,
		"provider":     c.Provider,
		"endpoint":     c.Endpoint,
		"service_name": c.ServiceName,
		"sample_rate":  c.SampleRate,
		"insecure":     c.Insecure,
		"headers":      c.Headers,
	}
}

// =============================================================================
// Circuit Breaker Options
// =============================================================================

// CircuitBreakerOptions returns circuit breaker configuration
func (c *CircuitBreakerConfig) CircuitBreakerOptions() map[string]interface{} {
	return map[string]interface{}{
		"enabled":         c.Enabled,
		"threshold":       c.Threshold,
		"timeout":         c.Timeout,
		"max_requests":    c.MaxRequests,
		"interval":        c.Interval,
		"on_state_change": c.OnStateChange,
	}
}

// =============================================================================
// Retry Options
// =============================================================================

// RetryOptions returns retry configuration
func (c *RetryConfig) RetryOptions() map[string]interface{} {
	return map[string]interface{}{
		"max_attempts":     c.MaxAttempts,
		"initial_interval": c.InitialInterval,
		"max_interval":     c.MaxInterval,
		"multiplier":       c.Multiplier,
		"randomize_factor": c.RandomizeFactor,
	}
}
