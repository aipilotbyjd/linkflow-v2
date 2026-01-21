package executor

import (
	"context"
	"sync"

	"github.com/google/uuid"
	billingapp "github.com/linkflow-ai/linkflow/internal/core/application/billing"
)

// UsageTracker tracks and enforces usage during workflow execution
type UsageTracker struct {
	usageService *billingapp.UsageService
	mu           sync.Mutex

	// Track operations consumed in current execution
	executionOps map[uuid.UUID]int64 // executionID -> operations consumed
}

func NewUsageTracker(usageService *billingapp.UsageService) *UsageTracker {
	return &UsageTracker{
		usageService: usageService,
		executionOps: make(map[uuid.UUID]int64),
	}
}

// PreExecutionCheck verifies workspace can execute (has operations available)
func (t *UsageTracker) PreExecutionCheck(ctx context.Context, workspaceID uuid.UUID, estimatedOps int64) error {
	if t.usageService == nil {
		return nil // Skip if no usage service configured
	}

	// Estimate at least 1 operation for the workflow trigger
	if estimatedOps < 1 {
		estimatedOps = 1
	}

	return t.usageService.CheckOperationsAvailable(ctx, workspaceID, estimatedOps)
}

// TrackNodeExecution records operation consumption for a node
func (t *UsageTracker) TrackNodeExecution(ctx context.Context, workspaceID, executionID uuid.UUID, nodeType string) error {
	if t.usageService == nil {
		return nil
	}

	// Each node execution = 1 operation (Make.com style)
	ops := t.getNodeOperationCost(nodeType)

	// Consume operations (pass nodeType for task-free check)
	if err := t.usageService.CheckAndConsumeOperations(ctx, workspaceID, ops, nodeType); err != nil {
		return err
	}

	// Track for this execution
	t.mu.Lock()
	t.executionOps[executionID] += ops
	t.mu.Unlock()

	return nil
}

// TrackAIUsage records AI credit consumption with token-based billing
func (t *UsageTracker) TrackAIUsage(ctx context.Context, workspaceID uuid.UUID, model string, inputTokens, outputTokens int) error {
	if t.usageService == nil {
		return nil
	}
	return t.usageService.CheckAndConsumeAICredits(ctx, workspaceID, model, inputTokens, outputTokens, 0, 0)
}

// TrackAIImageUsage records AI image generation credit consumption
func (t *UsageTracker) TrackAIImageUsage(ctx context.Context, workspaceID uuid.UUID, model string, images int) error {
	if t.usageService == nil {
		return nil
	}
	return t.usageService.CheckAndConsumeAICredits(ctx, workspaceID, model, 0, 0, images, 0)
}

// TrackAIAudioUsage records AI audio credit consumption
func (t *UsageTracker) TrackAIAudioUsage(ctx context.Context, workspaceID uuid.UUID, model string, minutes int) error {
	if t.usageService == nil {
		return nil
	}
	return t.usageService.CheckAndConsumeAICredits(ctx, workspaceID, model, 0, 0, 0, minutes)
}

// TrackDataTransfer records data transfer
func (t *UsageTracker) TrackDataTransfer(ctx context.Context, workspaceID uuid.UUID, bytes int64) error {
	if t.usageService == nil {
		return nil
	}
	return t.usageService.TrackDataTransfer(ctx, workspaceID, bytes)
}

// GetExecutionOperations returns operations consumed by an execution
func (t *UsageTracker) GetExecutionOperations(executionID uuid.UUID) int64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.executionOps[executionID]
}

// ClearExecution clears tracking for completed execution
func (t *UsageTracker) ClearExecution(executionID uuid.UUID) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.executionOps, executionID)
}

// getNodeOperationCost returns the operation cost for a node type
func (t *UsageTracker) getNodeOperationCost(nodeType string) int64 {
	// AI nodes cost more
	switch nodeType {
	case "action.openai", "action.anthropic", "action.ai_gateway",
		"action.ai_text", "action.ai_image", "action.ai_embedding":
		return 5 // AI operations cost 5x

	case "action.http", "action.webhook":
		return 1 // External calls

	case "logic.loop", "logic.batch":
		return 1 // Each iteration counts separately

	case "trigger.webhook", "trigger.schedule", "trigger.manual":
		return 1 // Triggers count as 1

	default:
		return 1 // Default: 1 operation per node
	}
}

// CalculateAICreditCost calculates AI credits for different AI operations
func CalculateAICreditCost(operation string, inputTokens, outputTokens int) int64 {
	// Credit costs based on token usage
	switch operation {
	case "gpt-4", "gpt-4-turbo":
		return int64((inputTokens/1000)*3 + (outputTokens/1000)*6)
	case "gpt-3.5-turbo":
		return int64((inputTokens/1000)*1 + (outputTokens/1000)*2)
	case "claude-3-opus":
		return int64((inputTokens/1000)*4 + (outputTokens/1000)*8)
	case "claude-3-sonnet":
		return int64((inputTokens/1000)*2 + (outputTokens/1000)*4)
	case "claude-3-haiku":
		return int64((inputTokens / 1000) + (outputTokens / 1000))
	case "dall-e-3":
		return 20 // Per image
	case "dall-e-2":
		return 5 // Per image
	case "whisper":
		return int64(inputTokens / 60) // Per minute of audio
	default:
		return int64((inputTokens + outputTokens) / 1000) // Default: 1 credit per 1000 tokens
	}
}
