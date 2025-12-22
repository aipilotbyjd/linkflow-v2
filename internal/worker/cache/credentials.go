package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/redis/go-redis/v9"
)

// CredentialCache caches decrypted credentials
type CredentialCache struct {
	redis    *redis.Client
	ttl      time.Duration
	memory   sync.Map // In-memory cache for hot credentials
	memTTL   time.Duration
}

// CredentialCacheConfig configures the credential cache
type CredentialCacheConfig struct {
	RedisTTL  time.Duration // TTL in Redis
	MemoryTTL time.Duration // TTL in memory (shorter for security)
}

// NewCredentialCache creates a new credential cache
func NewCredentialCache(redis *redis.Client, cfg CredentialCacheConfig) *CredentialCache {
	if cfg.RedisTTL == 0 {
		cfg.RedisTTL = 5 * time.Minute
	}
	if cfg.MemoryTTL == 0 {
		cfg.MemoryTTL = 1 * time.Minute
	}

	return &CredentialCache{
		redis:  redis,
		ttl:    cfg.RedisTTL,
		memTTL: cfg.MemoryTTL,
	}
}

// CachedCredential wraps credential data with expiry
type CachedCredential struct {
	Data      *models.CredentialData `json:"data"`
	CachedAt  time.Time              `json:"cached_at"`
	ExpiresAt time.Time              `json:"expires_at"`
}

// memoryCachedCredential is for in-memory caching
type memoryCachedCredential struct {
	Data      *models.CredentialData
	ExpiresAt time.Time
}

// Key generates a cache key for a credential
func (c *CredentialCache) Key(credentialID uuid.UUID) string {
	return fmt.Sprintf("credential:cache:%s", credentialID)
}

// Get retrieves a cached credential
func (c *CredentialCache) Get(ctx context.Context, credentialID uuid.UUID) (*models.CredentialData, bool) {
	// Check memory cache first
	if cached, ok := c.getFromMemory(credentialID); ok {
		return cached, true
	}

	// Check Redis
	key := c.Key(credentialID)
	data, err := c.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil, false
	}

	var cached CachedCredential
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, false
	}

	// Validate expiry
	if time.Now().After(cached.ExpiresAt) {
		c.redis.Del(ctx, key)
		return nil, false
	}

	// Store in memory cache
	c.setInMemory(credentialID, cached.Data)

	return cached.Data, true
}

// Set stores a credential in cache
func (c *CredentialCache) Set(ctx context.Context, credentialID uuid.UUID, data *models.CredentialData) error {
	cached := CachedCredential{
		Data:      data,
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(c.ttl),
	}

	jsonData, err := json.Marshal(cached)
	if err != nil {
		return err
	}

	key := c.Key(credentialID)
	if err := c.redis.Set(ctx, key, jsonData, c.ttl).Err(); err != nil {
		return err
	}

	// Store in memory cache
	c.setInMemory(credentialID, data)

	return nil
}

// Invalidate removes a credential from cache
func (c *CredentialCache) Invalidate(ctx context.Context, credentialID uuid.UUID) error {
	// Remove from memory
	c.memory.Delete(credentialID.String())

	// Remove from Redis
	key := c.Key(credentialID)
	return c.redis.Del(ctx, key).Err()
}

// InvalidateAll removes all credentials from cache
func (c *CredentialCache) InvalidateAll(ctx context.Context) error {
	// Clear memory cache
	c.memory = sync.Map{}

	// Clear Redis
	pattern := "credential:cache:*"
	iter := c.redis.Scan(ctx, 0, pattern, 100).Iterator()
	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return err
	}

	if len(keys) > 0 {
		return c.redis.Del(ctx, keys...).Err()
	}

	return nil
}

func (c *CredentialCache) getFromMemory(credentialID uuid.UUID) (*models.CredentialData, bool) {
	value, ok := c.memory.Load(credentialID.String())
	if !ok {
		return nil, false
	}

	cached := value.(*memoryCachedCredential)
	if time.Now().After(cached.ExpiresAt) {
		c.memory.Delete(credentialID.String())
		return nil, false
	}

	return cached.Data, true
}

func (c *CredentialCache) setInMemory(credentialID uuid.UUID, data *models.CredentialData) {
	cached := &memoryCachedCredential{
		Data:      data,
		ExpiresAt: time.Now().Add(c.memTTL),
	}
	c.memory.Store(credentialID.String(), cached)
}

// CleanupExpired removes expired entries from memory cache
func (c *CredentialCache) CleanupExpired() {
	now := time.Now()
	c.memory.Range(func(key, value interface{}) bool {
		cached := value.(*memoryCachedCredential)
		if now.After(cached.ExpiresAt) {
			c.memory.Delete(key)
		}
		return true
	})
}

// StartCleanupRoutine starts a background routine to clean up expired entries
func (c *CredentialCache) StartCleanupRoutine(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				c.CleanupExpired()
			}
		}
	}()
}

// CachedCredentialResolver creates a credential resolver with caching
type CachedCredentialResolver struct {
	cache   *CredentialCache
	fetcher func(ctx context.Context, id uuid.UUID) (*models.CredentialData, error)
}

// NewCachedCredentialResolver creates a cached credential resolver
func NewCachedCredentialResolver(cache *CredentialCache, fetcher func(ctx context.Context, id uuid.UUID) (*models.CredentialData, error)) *CachedCredentialResolver {
	return &CachedCredentialResolver{
		cache:   cache,
		fetcher: fetcher,
	}
}

// Get retrieves a credential, using cache when possible
func (r *CachedCredentialResolver) Get(ctx context.Context, credentialID uuid.UUID) (*models.CredentialData, error) {
	// Try cache first
	if data, ok := r.cache.Get(ctx, credentialID); ok {
		return data, nil
	}

	// Fetch from source
	data, err := r.fetcher(ctx, credentialID)
	if err != nil {
		return nil, err
	}

	// Store in cache
	_ = r.cache.Set(ctx, credentialID, data)

	return data, nil
}

// TokenRefreshCache caches OAuth token refresh results
type TokenRefreshCache struct {
	redis *redis.Client
	ttl   time.Duration
}

// NewTokenRefreshCache creates a token refresh cache
func NewTokenRefreshCache(redis *redis.Client, ttl time.Duration) *TokenRefreshCache {
	if ttl == 0 {
		ttl = 5 * time.Minute
	}
	return &TokenRefreshCache{redis: redis, ttl: ttl}
}

// RefreshLock represents a lock for token refresh
type RefreshLock struct {
	CredentialID uuid.UUID `json:"credential_id"`
	LockedAt     time.Time `json:"locked_at"`
	LockedBy     string    `json:"locked_by"`
}

// AcquireRefreshLock attempts to acquire a lock for token refresh
func (c *TokenRefreshCache) AcquireRefreshLock(ctx context.Context, credentialID uuid.UUID, lockedBy string) (bool, error) {
	key := fmt.Sprintf("credential:refresh:lock:%s", credentialID)

	lock := RefreshLock{
		CredentialID: credentialID,
		LockedAt:     time.Now(),
		LockedBy:     lockedBy,
	}

	data, err := json.Marshal(lock)
	if err != nil {
		return false, err
	}

	// Try to set with NX (only if not exists)
	result, err := c.redis.SetNX(ctx, key, data, 30*time.Second).Result()
	if err != nil {
		return false, err
	}

	return result, nil
}

// ReleaseRefreshLock releases a token refresh lock
func (c *TokenRefreshCache) ReleaseRefreshLock(ctx context.Context, credentialID uuid.UUID) error {
	key := fmt.Sprintf("credential:refresh:lock:%s", credentialID)
	return c.redis.Del(ctx, key).Err()
}

// CachedToken represents a cached OAuth token
type CachedToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// GetToken retrieves a cached token
func (c *TokenRefreshCache) GetToken(ctx context.Context, credentialID uuid.UUID) (*CachedToken, bool) {
	key := fmt.Sprintf("credential:token:%s", credentialID)
	data, err := c.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil, false
	}

	var token CachedToken
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, false
	}

	// Check expiry
	if time.Now().After(token.ExpiresAt) {
		c.redis.Del(ctx, key)
		return nil, false
	}

	return &token, true
}

// SetToken caches an OAuth token
func (c *TokenRefreshCache) SetToken(ctx context.Context, credentialID uuid.UUID, token *CachedToken) error {
	key := fmt.Sprintf("credential:token:%s", credentialID)

	data, err := json.Marshal(token)
	if err != nil {
		return err
	}

	// Calculate TTL based on token expiry
	ttl := time.Until(token.ExpiresAt)
	if ttl < 0 {
		ttl = c.ttl
	}

	return c.redis.Set(ctx, key, data, ttl).Err()
}

// InvalidateToken removes a cached token
func (c *TokenRefreshCache) InvalidateToken(ctx context.Context, credentialID uuid.UUID) error {
	key := fmt.Sprintf("credential:token:%s", credentialID)
	return c.redis.Del(ctx, key).Err()
}
