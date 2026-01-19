package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// AIUsage tracks AI API usage for billing and analytics
type AIUsage struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	WorkspaceID uuid.UUID  `gorm:"type:uuid;index;not null"`
	ExecutionID *uuid.UUID `gorm:"type:uuid;index"`
	NodeID      *string    `gorm:"size:255"`

	// Provider and model info
	Provider string `gorm:"size:50;not null;index"`
	Model    string `gorm:"size:100;not null;index"`

	// Request type: chat, completion, embedding, image, vision, tts, stt
	RequestType string `gorm:"size:50;not null"`

	// Token usage
	InputTokens  int `gorm:"not null;default:0"`
	OutputTokens int `gorm:"not null;default:0"`
	TotalTokens  int `gorm:"not null;default:0"`

	// Cost in USD
	CostUSD float64 `gorm:"type:decimal(10,6);not null;default:0"`

	// Performance
	LatencyMS int64 `gorm:"not null;default:0"`

	// Cache info
	Cached   bool `gorm:"not null;default:false"`
	CacheHit bool `gorm:"not null;default:false"`

	// Metadata
	Metadata  types.JSON `gorm:"type:jsonb"`
	CreatedAt time.Time  `gorm:"index;default:now()"`

	// Relations
	Workspace Workspace  `gorm:"foreignKey:WorkspaceID"`
	Execution *Execution `gorm:"foreignKey:ExecutionID"`
}

func (AIUsage) TableName() string {
	return "ai_usage"
}

// PromptTemplate represents a reusable prompt template
type PromptTemplate struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	WorkspaceID uuid.UUID `gorm:"type:uuid;index;not null"`
	CreatedBy   uuid.UUID `gorm:"type:uuid;index;not null"`

	// Template info
	Name        string  `gorm:"size:255;not null"`
	Description *string `gorm:"type:text"`
	Category    *string `gorm:"size:100;index"`

	// Template content
	Template  string     `gorm:"type:text;not null"`
	Variables types.JSON `gorm:"type:jsonb;not null;default:'[]'"`

	// System message
	SystemMessage *string `gorm:"type:text"`

	// Default model settings
	DefaultModel       *string  `gorm:"size:100"`
	DefaultTemperature *float64 `gorm:"type:decimal(3,2)"`
	DefaultMaxTokens   *int

	// Versioning
	Version  int  `gorm:"not null;default:1"`
	IsActive bool `gorm:"not null;default:true"`
	IsPublic bool `gorm:"not null;default:false;index"`

	// Metadata
	Tags      pq.StringArray `gorm:"type:text[]"`
	CreatedAt time.Time      `gorm:"default:now()"`
	UpdatedAt time.Time      `gorm:"default:now()"`

	// Relations
	Workspace Workspace `gorm:"foreignKey:WorkspaceID"`
	Creator   User      `gorm:"foreignKey:CreatedBy"`
}

func (PromptTemplate) TableName() string {
	return "prompt_templates"
}

// AICache stores cached AI responses
type AICache struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	WorkspaceID *uuid.UUID `gorm:"type:uuid;index"`

	// Cache key
	PromptHash string `gorm:"size:64;not null;index"`
	Model      string `gorm:"size:100;not null;index"`
	Provider   string `gorm:"size:50;not null"`

	// Cached response
	Response    types.JSON `gorm:"type:jsonb;not null"`
	RequestType string     `gorm:"size:50;not null"`

	// Metadata
	HitCount   int       `gorm:"not null;default:0"`
	CreatedAt  time.Time `gorm:"default:now()"`
	ExpiresAt  time.Time `gorm:"index;not null"`
	LastUsedAt time.Time `gorm:"default:now()"`

	// Relations
	Workspace *Workspace `gorm:"foreignKey:WorkspaceID"`
}

func (AICache) TableName() string {
	return "ai_cache"
}

// Indexes for AICache (composite unique index)
func (AICache) Indexes() []string {
	return []string{
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_ai_cache_unique_prompt ON ai_cache(prompt_hash, model, workspace_id)",
	}
}
