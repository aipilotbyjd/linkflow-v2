package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

// TokenRefreshPayload contains data for OAuth token refresh task
type TokenRefreshPayload struct {
	CredentialID string `json:"credential_id"`
	WorkspaceID  string `json:"workspace_id"`
	Provider     string `json:"provider"`
}

// NewTokenRefreshTask creates a new token refresh task
func NewTokenRefreshTask(payload TokenRefreshPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	return asynq.NewTask(
		TypeTokenRefresh,
		data,
		asynq.MaxRetry(3),
		asynq.Timeout(1*time.Minute),
		asynq.Queue(QueueCritical),
	), nil
}

// TokenRefreshHandler handles token refresh tasks
type TokenRefreshHandler struct {
	refresher TokenRefresher
}

// TokenRefresher interface for refreshing OAuth tokens
type TokenRefresher interface {
	RefreshToken(ctx context.Context, credentialID, workspaceID, provider string) error
}

// NewTokenRefreshHandler creates a new handler
func NewTokenRefreshHandler(refresher TokenRefresher) *TokenRefreshHandler {
	return &TokenRefreshHandler{refresher: refresher}
}

// ProcessTask processes a token refresh task
func (h *TokenRefreshHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	var payload TokenRefreshPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return h.refresher.RefreshToken(ctx, payload.CredentialID, payload.WorkspaceID, payload.Provider)
}
