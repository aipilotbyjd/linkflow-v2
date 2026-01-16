package middleware

import (
	"context"

	"github.com/hibiken/asynq"
)

type traceContextKey struct{}

// TraceInfo contains trace information for a task
type TraceInfo struct {
	TraceID string
	SpanID  string
	TaskID  string
	TaskType string
}

// Tracing returns a middleware that adds trace context to tasks
func Tracing() asynq.MiddlewareFunc {
	return func(h asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			taskID, _ := asynq.GetTaskID(ctx)

			traceInfo := &TraceInfo{
				TraceID:  generateID(),
				SpanID:   generateID(),
				TaskID:   taskID,
				TaskType: t.Type(),
			}

			ctx = context.WithValue(ctx, traceContextKey{}, traceInfo)
			return h.ProcessTask(ctx, t)
		})
	}
}

// GetTraceInfo retrieves trace info from context
func GetTraceInfo(ctx context.Context) *TraceInfo {
	if info, ok := ctx.Value(traceContextKey{}).(*TraceInfo); ok {
		return info
	}
	return nil
}

func generateID() string {
	return randomString(16)
}

func randomString(n int) string {
	const letters = "abcdef0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[i%len(letters)]
	}
	return string(b)
}
