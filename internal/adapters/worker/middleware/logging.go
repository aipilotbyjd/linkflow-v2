package middleware

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

// Logging returns a middleware that logs task execution
func Logging() asynq.MiddlewareFunc {
	return func(h asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			start := time.Now()
			taskID, _ := asynq.GetTaskID(ctx)

			log.Info().
				Str("task_type", t.Type()).
				Str("task_id", taskID).
				Int("payload_size", len(t.Payload())).
				Msg("Task started")

			err := h.ProcessTask(ctx, t)

			duration := time.Since(start)

			if err != nil {
				log.Error().
					Err(err).
					Str("task_type", t.Type()).
					Str("task_id", taskID).
					Dur("duration", duration).
					Msg("Task failed")
			} else {
				log.Info().
					Str("task_type", t.Type()).
					Str("task_id", taskID).
					Dur("duration", duration).
					Msg("Task completed")
			}

			return err
		})
	}
}
