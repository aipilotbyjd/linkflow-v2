package logic

import (
	"context"
	"time"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type WaitNode struct{}

func NewWaitNode() *WaitNode {
	return &WaitNode{}
}

func (n *WaitNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})

	seconds := 1.0
	if s, ok := params["seconds"].(float64); ok {
		seconds = s
	}

	select {
	case <-time.After(time.Duration(seconds) * time.Second):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	return types.JSON{
		"waited_seconds": seconds,
	}, nil
}

func (n *WaitNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "logic.wait",
		Name:        "Wait",
		Description: "Pause workflow execution for a specified duration or until a specific time",
		Category:    "logic",
		Version:     "1.0.0",
		Icon:        "Clock02",
		Color:       "#64748B",
		Inputs: []wtypes.NodePort{
			{Name: "main", Type: "any", Description: "Input data (passed through after wait)"},
		},
		Outputs: []wtypes.NodePort{
			{Name: "main", Type: "any", Description: "Same data after wait completes"},
		},
		Parameters: []wtypes.NodeParameter{
			{
				Name:        "mode",
				DisplayName: "Wait Mode",
				Type:        "options",
				Required:    true,
				Default:     "duration",
				Description: "How to determine wait time",
				Options: []wtypes.ParamOption{
					{Name: "Duration", Value: "duration", Description: "Wait for a specific amount of time"},
					{Name: "Until Time", Value: "until", Description: "Wait until a specific date/time"},
					{Name: "Webhook", Value: "webhook", Description: "Wait for an external webhook call"},
				},
			},
			{
				Name:        "duration_value",
				DisplayName: "Duration",
				Type:        "number",
				Required:    true,
				Default:     5,
				Description: "Amount of time to wait",
				ShowIf:      "mode === 'duration'",
			},
			{
				Name:        "duration_unit",
				DisplayName: "Unit",
				Type:        "options",
				Required:    true,
				Default:     "seconds",
				Description: "Time unit",
				ShowIf:      "mode === 'duration'",
				Options: []wtypes.ParamOption{
					{Name: "Milliseconds", Value: "milliseconds"},
					{Name: "Seconds", Value: "seconds"},
					{Name: "Minutes", Value: "minutes"},
					{Name: "Hours", Value: "hours"},
					{Name: "Days", Value: "days"},
				},
			},
			{
				Name:        "until_time",
				DisplayName: "Wait Until",
				Type:        "string",
				Required:    true,
				Description: "Date/time to wait until (ISO 8601 format or expression)",
				Placeholder: "2024-12-31T23:59:59Z",
				ShowIf:      "mode === 'until'",
			},
			{
				Name:        "timeout",
				DisplayName: "Timeout (seconds)",
				Type:        "number",
				Required:    false,
				Default:     3600,
				Description: "Maximum time to wait for webhook (in seconds)",
				ShowIf:      "mode === 'webhook'",
			},
			{
				Name:        "seconds",
				DisplayName: "Seconds (Legacy)",
				Type:        "number",
				Required:    false,
				Default:     1,
				Description: "Seconds to wait (legacy parameter)",
			},
		},
	}
}
