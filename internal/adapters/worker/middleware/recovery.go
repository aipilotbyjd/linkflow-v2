package middleware

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

// Recovery returns a middleware that recovers from panics
func Recovery() asynq.MiddlewareFunc {
	return func(h asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) (err error) {
			defer func() {
				if r := recover(); r != nil {
					taskID, _ := asynq.GetTaskID(ctx)
					stack := debug.Stack()

					log.Error().
						Str("task_type", t.Type()).
						Str("task_id", taskID).
						Interface("panic", r).
						Str("stack", string(stack)).
						Msg("Task panicked")

					err = fmt.Errorf("task panicked: %v", r)
				}
			}()

			return h.ProcessTask(ctx, t)
		})
	}
}
