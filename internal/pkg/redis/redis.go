package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/linkflow-ai/linkflow/internal/pkg/config"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type Client struct {
	*redis.Client
}

func NewClient(cfg *config.RedisConfig) (*Client, error) {
	opts := &redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	}

	if cfg.TLS {
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	log.Info().Str("addr", cfg.Addr()).Msg("Redis connected successfully")

	return &Client{client}, nil
}

// Cache operations
func (c *Client) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.Set(ctx, key, value, expiration).Err()
}

func (c *Client) GetJSON(ctx context.Context, key string, dest interface{}) error {
	return c.Get(ctx, key).Scan(dest)
}

// Rate limiting
func (c *Client) RateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, int, error) {
	pipe := c.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, err
	}

	count := int(incr.Val())
	return count <= limit, limit - count, nil
}

// Pub/Sub
func (c *Client) PublishEvent(ctx context.Context, channel string, message interface{}) error {
	return c.Publish(ctx, channel, message).Err()
}

func (c *Client) SubscribeToChannel(ctx context.Context, channel string) *redis.PubSub {
	return c.Subscribe(ctx, channel)
}

// Leader election
func (c *Client) AcquireLock(ctx context.Context, key string, value string, ttl time.Duration) (bool, error) {
	return c.SetNX(ctx, key, value, ttl).Result()
}

func (c *Client) ReleaseLock(ctx context.Context, key string, value string) error {
	script := redis.NewScript(`
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`)
	return script.Run(ctx, c.Client, []string{key}, value).Err()
}

func (c *Client) ExtendLock(ctx context.Context, key string, value string, ttl time.Duration) (bool, error) {
	script := redis.NewScript(`
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("pexpire", KEYS[1], ARGV[2])
		else
			return 0
		end
	`)
	result, err := script.Run(ctx, c.Client, []string{key}, value, ttl.Milliseconds()).Int()
	return result == 1, err
}

// Idempotency
func (c *Client) CheckIdempotency(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return c.SetNX(ctx, "idempotency:"+key, "1", ttl).Result()
}

func (c *Client) GetIdempotencyResult(ctx context.Context, key string) (string, error) {
	return c.Get(ctx, "idempotency:result:"+key).Result()
}

func (c *Client) SetIdempotencyResult(ctx context.Context, key string, result string, ttl time.Duration) error {
	return c.Set(ctx, "idempotency:result:"+key, result, ttl).Err()
}

// Token blacklist operations
const tokenBlacklistPrefix = "token:blacklist:"

// BlacklistToken adds a token to the blacklist with TTL matching token expiry
func (c *Client) BlacklistToken(ctx context.Context, jti string, expiry time.Duration) error {
	return c.Set(ctx, tokenBlacklistPrefix+jti, "1", expiry).Err()
}

// IsTokenBlacklisted checks if a token is in the blacklist
func (c *Client) IsTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	result, err := c.Exists(ctx, tokenBlacklistPrefix+jti).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// BlacklistUserTokens blacklists all tokens for a user by storing user ID with timestamp
// Any token issued before this timestamp should be considered invalid
func (c *Client) BlacklistUserTokens(ctx context.Context, userID string, expiry time.Duration) error {
	return c.Set(ctx, "user:logout:"+userID, time.Now().Unix(), expiry).Err()
}

// GetUserLogoutTime returns the timestamp when user logged out (0 if never)
func (c *Client) GetUserLogoutTime(ctx context.Context, userID string) (int64, error) {
	result, err := c.Get(ctx, "user:logout:"+userID).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return result, err
}
