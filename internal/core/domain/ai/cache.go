package ai

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// CacheEntry represents a cached AI response
type CacheEntry struct {
	ID          uuid.UUID `json:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id,omitempty"`

	// Cache key
	PromptHash string   `json:"prompt_hash"`
	Model      string   `json:"model"`
	Provider   Provider `json:"provider"`

	// For semantic cache
	Embedding []float64 `json:"embedding,omitempty"`

	// Cached response
	Response    []byte `json:"response"`
	RequestType string `json:"request_type"` // chat, completion, embedding

	// Metadata
	HitCount   int       `json:"hit_count"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	LastUsedAt time.Time `json:"last_used_at"`
}

// IsExpired checks if the cache entry has expired
func (c *CacheEntry) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// IncrementHitCount increments the hit count
func (c *CacheEntry) IncrementHitCount() {
	c.HitCount++
	c.LastUsedAt = time.Now()
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	// Enable/disable caching
	Enabled bool `json:"enabled"`

	// TTL for cache entries
	TTL time.Duration `json:"ttl"`

	// Semantic cache settings
	SemanticEnabled     bool    `json:"semantic_enabled"`
	SimilarityThreshold float64 `json:"similarity_threshold"` // 0.0 to 1.0

	// Max entries per workspace
	MaxEntriesPerWorkspace int `json:"max_entries_per_workspace"`
}

// DefaultCacheConfig returns default cache configuration
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		Enabled:                true,
		TTL:                    24 * time.Hour,
		SemanticEnabled:        true,
		SimilarityThreshold:    0.95,
		MaxEntriesPerWorkspace: 10000,
	}
}

// Cache defines the interface for AI response caching
type Cache interface {
	// Get retrieves a cached response by exact hash match
	Get(ctx context.Context, hash, model string) (*CacheEntry, error)

	// GetSemantic retrieves a cached response by semantic similarity
	GetSemantic(ctx context.Context, embedding []float64, model string, threshold float64) (*CacheEntry, error)

	// Set stores a response in the cache
	Set(ctx context.Context, entry *CacheEntry) error

	// Delete removes a cache entry
	Delete(ctx context.Context, id uuid.UUID) error

	// Clear clears all cache entries for a workspace
	Clear(ctx context.Context, workspaceID uuid.UUID) error

	// Cleanup removes expired entries
	Cleanup(ctx context.Context) (int, error)

	// Stats returns cache statistics
	Stats(ctx context.Context, workspaceID uuid.UUID) (*CacheStats, error)
}

// CacheStats represents cache statistics
type CacheStats struct {
	TotalEntries     int       `json:"total_entries"`
	TotalHits        int       `json:"total_hits"`
	TotalMisses      int       `json:"total_misses"`
	HitRate          float64   `json:"hit_rate"`
	EstimatedSavings float64   `json:"estimated_savings_usd"`
	OldestEntry      time.Time `json:"oldest_entry"`
	NewestEntry      time.Time `json:"newest_entry"`
}

// CacheKeyGenerator generates cache keys
type CacheKeyGenerator struct{}

// GenerateHash generates a hash for cache lookup
func (g *CacheKeyGenerator) GenerateHash(messages []Message, model string, temperature *float64) string {
	// Implementation would hash the messages and parameters
	// For now, return a placeholder
	return ""
}

// NewCacheKeyGenerator creates a new cache key generator
func NewCacheKeyGenerator() *CacheKeyGenerator {
	return &CacheKeyGenerator{}
}
