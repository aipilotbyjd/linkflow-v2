package ai

import (
	"time"

	"github.com/google/uuid"
)

// UsageRecord represents a single AI usage record for tracking
type UsageRecord struct {
	ID          uuid.UUID `json:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	ExecutionID uuid.UUID `json:"execution_id,omitempty"`
	NodeID      string    `json:"node_id,omitempty"`

	// Provider and model info
	Provider Provider `json:"provider"`
	Model    string   `json:"model"`

	// Request type
	RequestType string `json:"request_type"` // chat, completion, embedding, image, vision, tts, stt

	// Token usage
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`

	// Cost
	CostUSD float64 `json:"cost_usd"`

	// Performance
	LatencyMS int64 `json:"latency_ms"`

	// Cache info
	Cached   bool `json:"cached"`
	CacheHit bool `json:"cache_hit"`

	// Metadata
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

// NewUsageRecord creates a new usage record
func NewUsageRecord(workspaceID uuid.UUID, provider Provider, model, requestType string) *UsageRecord {
	return &UsageRecord{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		Provider:    provider,
		Model:       model,
		RequestType: requestType,
		CreatedAt:   time.Now(),
		Metadata:    make(map[string]string),
	}
}

// SetUsage sets the token usage
func (u *UsageRecord) SetUsage(input, output int) {
	u.InputTokens = input
	u.OutputTokens = output
	u.TotalTokens = input + output
}

// CalculateCost calculates the cost based on the model pricing
func (u *UsageRecord) CalculateCost() {
	model, ok := GetModel(u.Model)
	if !ok {
		return
	}

	inputCost := float64(u.InputTokens) * model.InputPricePer1M / 1_000_000
	outputCost := float64(u.OutputTokens) * model.OutputPricePer1M / 1_000_000
	u.CostUSD = inputCost + outputCost
}

// UsageSummary represents aggregated usage statistics
type UsageSummary struct {
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Period      string    `json:"period"` // day, week, month

	// Totals
	TotalRequests     int     `json:"total_requests"`
	TotalTokens       int     `json:"total_tokens"`
	TotalInputTokens  int     `json:"total_input_tokens"`
	TotalOutputTokens int     `json:"total_output_tokens"`
	TotalCostUSD      float64 `json:"total_cost_usd"`

	// By provider
	ByProvider map[Provider]ProviderUsage `json:"by_provider"`

	// By model
	ByModel map[string]ModelUsage `json:"by_model"`

	// By request type
	ByRequestType map[string]int `json:"by_request_type"`

	// Time range
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// ProviderUsage represents usage for a specific provider
type ProviderUsage struct {
	Requests     int     `json:"requests"`
	Tokens       int     `json:"tokens"`
	CostUSD      float64 `json:"cost_usd"`
	AvgLatencyMS int64   `json:"avg_latency_ms"`
}

// ModelUsage represents usage for a specific model
type ModelUsage struct {
	Requests     int     `json:"requests"`
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	CostUSD      float64 `json:"cost_usd"`
	AvgLatencyMS int64   `json:"avg_latency_ms"`
}

// UsageRepository defines the interface for usage data persistence
type UsageRepository interface {
	// Create creates a new usage record
	Create(record *UsageRecord) error

	// GetByWorkspace returns usage records for a workspace
	GetByWorkspace(workspaceID uuid.UUID, start, end time.Time, limit int) ([]UsageRecord, error)

	// GetByExecution returns usage records for an execution
	GetByExecution(executionID uuid.UUID) ([]UsageRecord, error)

	// GetSummary returns aggregated usage statistics
	GetSummary(workspaceID uuid.UUID, period string, start, end time.Time) (*UsageSummary, error)

	// GetTotalCost returns total cost for a workspace in a time range
	GetTotalCost(workspaceID uuid.UUID, start, end time.Time) (float64, error)
}
