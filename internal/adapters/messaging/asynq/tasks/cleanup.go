package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

// CleanupPayload contains data for cleanup task
type CleanupPayload struct {
	Type        string    `json:"type"` // executions, logs, sessions, etc.
	OlderThan   time.Time `json:"older_than"`
	WorkspaceID string    `json:"workspace_id,omitempty"` // Optional: limit to specific workspace
	Limit       int       `json:"limit,omitempty"`        // Optional: max items to delete
}

// Cleanup types
const (
	CleanupTypeExecutions = "executions"
	CleanupTypeLogs       = "logs"
	CleanupTypeSessions   = "sessions"
	CleanupTypeBinaryData = "binary_data"
	CleanupTypeAll        = "all"
)

// NewCleanupTask creates a new cleanup task
func NewCleanupTask(payload CleanupPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	return asynq.NewTask(
		TypeCleanup,
		data,
		asynq.MaxRetry(1),
		asynq.Timeout(1*time.Hour),
		asynq.Queue(QueueLow),
	), nil
}

// CleanupHandler handles cleanup tasks
type CleanupHandler struct {
	cleaner DataCleaner
}

// DataCleaner interface for cleaning up old data
type DataCleaner interface {
	CleanExecutions(ctx context.Context, olderThan time.Time, workspaceID string, limit int) (int64, error)
	CleanLogs(ctx context.Context, olderThan time.Time, workspaceID string, limit int) (int64, error)
	CleanSessions(ctx context.Context, olderThan time.Time, limit int) (int64, error)
	CleanBinaryData(ctx context.Context, olderThan time.Time, workspaceID string, limit int) (int64, error)
}

// NewCleanupHandler creates a new handler
func NewCleanupHandler(cleaner DataCleaner) *CleanupHandler {
	return &CleanupHandler{cleaner: cleaner}
}

// ProcessTask processes a cleanup task
func (h *CleanupHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	var payload CleanupPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	switch payload.Type {
	case CleanupTypeExecutions:
		_, err := h.cleaner.CleanExecutions(ctx, payload.OlderThan, payload.WorkspaceID, payload.Limit)
		return err
	case CleanupTypeLogs:
		_, err := h.cleaner.CleanLogs(ctx, payload.OlderThan, payload.WorkspaceID, payload.Limit)
		return err
	case CleanupTypeSessions:
		_, err := h.cleaner.CleanSessions(ctx, payload.OlderThan, payload.Limit)
		return err
	case CleanupTypeBinaryData:
		_, err := h.cleaner.CleanBinaryData(ctx, payload.OlderThan, payload.WorkspaceID, payload.Limit)
		return err
	case CleanupTypeAll:
		if _, err := h.cleaner.CleanExecutions(ctx, payload.OlderThan, payload.WorkspaceID, payload.Limit); err != nil {
			return err
		}
		if _, err := h.cleaner.CleanLogs(ctx, payload.OlderThan, payload.WorkspaceID, payload.Limit); err != nil {
			return err
		}
		if _, err := h.cleaner.CleanSessions(ctx, payload.OlderThan, payload.Limit); err != nil {
			return err
		}
		_, err := h.cleaner.CleanBinaryData(ctx, payload.OlderThan, payload.WorkspaceID, payload.Limit)
		return err
	default:
		return fmt.Errorf("unknown cleanup type: %s", payload.Type)
	}
}
