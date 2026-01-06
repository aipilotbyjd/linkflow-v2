package tasks

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/rs/zerolog/log"
)

const (
	// TypeTokenRefresh is the task type for OAuth token refresh
	TypeTokenRefresh = "oauth:token_refresh"

	// DefaultRefreshInterval is how often to run token refresh
	DefaultRefreshInterval = 30 * time.Minute

	// DefaultRefreshWithin is how far in advance to refresh tokens
	DefaultRefreshWithin = 1 * time.Hour
)

// TokenRefreshPayload contains the payload for token refresh task
type TokenRefreshPayload struct {
	RefreshWithin time.Duration `json:"refresh_within"`
}

// TokenRefreshHandler handles OAuth token refresh tasks
type TokenRefreshHandler struct {
	oauthSvc *services.OAuthService
}

// NewTokenRefreshHandler creates a new token refresh handler
func NewTokenRefreshHandler(oauthSvc *services.OAuthService) *TokenRefreshHandler {
	return &TokenRefreshHandler{
		oauthSvc: oauthSvc,
	}
}

// Handle processes the token refresh task
func (h *TokenRefreshHandler) Handle(ctx context.Context, task *asynq.Task) error {
	var payload TokenRefreshPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		// Default if no payload
		payload.RefreshWithin = DefaultRefreshWithin
	}

	if payload.RefreshWithin == 0 {
		payload.RefreshWithin = DefaultRefreshWithin
	}

	log.Info().
		Dur("refresh_within", payload.RefreshWithin).
		Msg("Starting OAuth token refresh task")

	refreshed, failed, err := h.oauthSvc.RefreshExpiringTokens(ctx, payload.RefreshWithin)
	if err != nil {
		log.Error().Err(err).Msg("Token refresh task failed")
		return err
	}

	log.Info().
		Int("refreshed", refreshed).
		Int("failed", failed).
		Msg("OAuth token refresh task completed")

	return nil
}

// NewTokenRefreshTask creates a new token refresh task
func NewTokenRefreshTask(refreshWithin time.Duration) (*asynq.Task, error) {
	payload, err := json.Marshal(TokenRefreshPayload{
		RefreshWithin: refreshWithin,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeTokenRefresh, payload), nil
}

// TokenRefreshScheduler schedules periodic token refresh tasks
type TokenRefreshScheduler struct {
	client   *asynq.Client
	interval time.Duration
	within   time.Duration
}

// NewTokenRefreshScheduler creates a new scheduler
func NewTokenRefreshScheduler(redisAddr string, interval, within time.Duration) *TokenRefreshScheduler {
	if interval == 0 {
		interval = DefaultRefreshInterval
	}
	if within == 0 {
		within = DefaultRefreshWithin
	}

	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})

	return &TokenRefreshScheduler{
		client:   client,
		interval: interval,
		within:   within,
	}
}

// Start begins scheduling token refresh tasks
func (s *TokenRefreshScheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	log.Info().
		Dur("interval", s.interval).
		Dur("refresh_within", s.within).
		Msg("Starting OAuth token refresh scheduler")

	// Run immediately on start
	s.enqueueTask()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Token refresh scheduler stopped")
			return
		case <-ticker.C:
			s.enqueueTask()
		}
	}
}

func (s *TokenRefreshScheduler) enqueueTask() {
	task, err := NewTokenRefreshTask(s.within)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create token refresh task")
		return
	}

	info, err := s.client.Enqueue(task, asynq.MaxRetry(3), asynq.Timeout(5*time.Minute))
	if err != nil {
		log.Error().Err(err).Msg("Failed to enqueue token refresh task")
		return
	}

	log.Debug().
		Str("task_id", info.ID).
		Str("queue", info.Queue).
		Msg("Token refresh task enqueued")
}

// Close closes the scheduler client
func (s *TokenRefreshScheduler) Close() error {
	return s.client.Close()
}
