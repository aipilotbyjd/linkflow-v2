package jwt

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	blacklistPrefix = "token:blacklist:"
)

// Blacklist handles token blacklisting using Redis
type Blacklist struct {
	client *redis.Client
}

// NewBlacklist creates a new token blacklist
func NewBlacklist(client *redis.Client) *Blacklist {
	return &Blacklist{client: client}
}

// Add adds a token to the blacklist
func (b *Blacklist) Add(ctx context.Context, tokenID string, expiry time.Duration) error {
	key := blacklistPrefix + tokenID
	return b.client.Set(ctx, key, "1", expiry).Err()
}

// AddWithExpiration adds a token to the blacklist with specific expiration time
func (b *Blacklist) AddWithExpiration(ctx context.Context, tokenID string, expiresAt time.Time) error {
	key := blacklistPrefix + tokenID
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil // Token already expired
	}
	return b.client.Set(ctx, key, "1", ttl).Err()
}

// IsBlacklisted checks if a token is blacklisted
func (b *Blacklist) IsBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	key := blacklistPrefix + tokenID
	exists, err := b.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check blacklist: %w", err)
	}
	return exists > 0, nil
}

// Remove removes a token from the blacklist
func (b *Blacklist) Remove(ctx context.Context, tokenID string) error {
	key := blacklistPrefix + tokenID
	return b.client.Del(ctx, key).Err()
}

// BlacklistUser blacklists all tokens for a user by adding a user-level entry
func (b *Blacklist) BlacklistUser(ctx context.Context, userID string, expiry time.Duration) error {
	key := blacklistPrefix + "user:" + userID
	return b.client.Set(ctx, key, time.Now().Unix(), expiry).Err()
}

// IsUserBlacklisted checks if all user tokens are blacklisted
func (b *Blacklist) IsUserBlacklisted(ctx context.Context, userID string, tokenIssuedAt time.Time) (bool, error) {
	key := blacklistPrefix + "user:" + userID
	val, err := b.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check user blacklist: %w", err)
	}
	// Token is blacklisted if it was issued before the blacklist timestamp
	return tokenIssuedAt.Unix() < val, nil
}

// Cleanup removes expired entries (handled automatically by Redis TTL)
func (b *Blacklist) Cleanup(ctx context.Context) error {
	// Redis handles TTL cleanup automatically
	return nil
}

// Count returns the number of blacklisted tokens (for monitoring)
func (b *Blacklist) Count(ctx context.Context) (int64, error) {
	keys, err := b.client.Keys(ctx, blacklistPrefix+"*").Result()
	if err != nil {
		return 0, fmt.Errorf("failed to count blacklist: %w", err)
	}
	return int64(len(keys)), nil
}
