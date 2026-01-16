package middleware

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

// RetryConfig holds retry middleware configuration
type RetryConfig struct {
	MaxRetries int
	RetryTypes []string
}

// Retry returns a middleware that handles retry logic
func Retry(cfg RetryConfig) asynq.MiddlewareFunc {
	retryableTypes := make(map[string]bool)
	for _, t := range cfg.RetryTypes {
		retryableTypes[t] = true
	}

	return func(h asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			err := h.ProcessTask(ctx, t)
			if err == nil {
				return nil
			}

			taskID, _ := asynq.GetTaskID(ctx)
			retried, _ := asynq.GetRetryCount(ctx)
			maxRetry, _ := asynq.GetMaxRetry(ctx)

			if len(retryableTypes) > 0 && !retryableTypes[t.Type()] {
				log.Warn().
					Str("task_type", t.Type()).
					Str("task_id", taskID).
					Err(err).
					Msg("Task failed, not retryable type")
				return err
			}

			if retried < maxRetry {
				log.Warn().
					Str("task_type", t.Type()).
					Str("task_id", taskID).
					Int("retry_count", retried).
					Int("max_retry", maxRetry).
					Err(err).
					Msg("Task failed, will retry")
			} else {
				log.Error().
					Str("task_type", t.Type()).
					Str("task_id", taskID).
					Int("retry_count", retried).
					Err(err).
					Msg("Task exhausted all retries")
			}

			return err
		})
	}
}
