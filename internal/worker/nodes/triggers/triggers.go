package triggers

import (
	"context"
	"time"

	"github.com/linkflow-ai/linkflow/internal/worker/core"
)

func init() {
	// Auto-register all trigger nodes
	core.Register(&ManualTrigger{}, core.NodeMeta{
		Name:        "Manual Trigger",
		Description: "Manually trigger a workflow",
		Category:    "triggers",
		Icon:        "play",
		Version:     "1.0.0",
	})

	core.Register(&WebhookTrigger{}, core.NodeMeta{
		Name:        "Webhook Trigger",
		Description: "Trigger workflow via HTTP webhook",
		Category:    "triggers",
		Icon:        "webhook",
		Version:     "1.0.0",
	})

	core.Register(&ScheduleTrigger{}, core.NodeMeta{
		Name:        "Schedule Trigger",
		Description: "Trigger workflow on a schedule (cron)",
		Category:    "triggers",
		Icon:        "clock",
		Version:     "1.0.0",
	})
}

// ManualTrigger starts workflow manually
type ManualTrigger struct{}

func (n *ManualTrigger) Type() string { return "trigger.manual" }

func (n *ManualTrigger) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	return map[string]interface{}{
		"triggered": true,
		"input":     execCtx.Input["$input"],
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// WebhookTrigger starts workflow from HTTP request
type WebhookTrigger struct{}

func (n *WebhookTrigger) Type() string { return "trigger.webhook" }

func (n *WebhookTrigger) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	return map[string]interface{}{
		"triggered": true,
		"headers":   execCtx.Input["headers"],
		"body":      execCtx.Input["body"],
		"query":     execCtx.Input["query"],
		"method":    execCtx.Input["method"],
		"data":      execCtx.Input["$input"],
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// ScheduleTrigger starts workflow on schedule
type ScheduleTrigger struct{}

func (n *ScheduleTrigger) Type() string { return "trigger.schedule" }

func (n *ScheduleTrigger) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	return map[string]interface{}{
		"triggered":     true,
		"scheduledTime": execCtx.Input["scheduledTime"],
		"input":         execCtx.Input["$input"],
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
	}, nil
}
