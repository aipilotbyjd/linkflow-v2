package triggers

import (
	"context"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type ScheduleTrigger struct{}

func NewScheduleTrigger() *ScheduleTrigger {
	return &ScheduleTrigger{}
}

func (t *ScheduleTrigger) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	return runtime.GetInputData(), nil
}

func (t *ScheduleTrigger) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "trigger.schedule",
		Name:        "Schedule Trigger",
		Description: "Trigger workflow automatically on a recurring schedule using cron expressions or intervals",
		Category:    "trigger",
		Version:     "1.0.0",
		Icon:        "Clock01",
		Color:       "#8B5CF6",
		Inputs:      []wtypes.NodePort{},
		Outputs: []wtypes.NodePort{
			{Name: "main", Type: "object", Description: "Schedule execution context with timestamp and metadata"},
		},
		Parameters: []wtypes.NodeParameter{
			{
				Name:        "schedule_type",
				DisplayName: "Schedule Type",
				Type:        "options",
				Required:    true,
				Default:     "interval",
				Description: "How to define the schedule",
				Options: []wtypes.ParamOption{
					{Name: "Interval", Value: "interval", Description: "Run at fixed intervals"},
					{Name: "Cron Expression", Value: "cron", Description: "Use cron syntax for complex schedules"},
				},
			},
			{
				Name:        "interval_value",
				DisplayName: "Interval",
				Type:        "number",
				Required:    true,
				Default:     5,
				Description: "Time between executions",
				ShowIf:      "schedule_type === 'interval'",
			},
			{
				Name:        "interval_unit",
				DisplayName: "Unit",
				Type:        "options",
				Required:    true,
				Default:     "minutes",
				Description: "Time unit for the interval",
				ShowIf:      "schedule_type === 'interval'",
				Options: []wtypes.ParamOption{
					{Name: "Seconds", Value: "seconds"},
					{Name: "Minutes", Value: "minutes"},
					{Name: "Hours", Value: "hours"},
					{Name: "Days", Value: "days"},
					{Name: "Weeks", Value: "weeks"},
				},
			},
			{
				Name:        "cron_expression",
				DisplayName: "Cron Expression",
				Type:        "string",
				Required:    true,
				Description: "Cron expression (e.g., '0 9 * * 1-5' for 9 AM weekdays)",
				Placeholder: "0 * * * *",
				ShowIf:      "schedule_type === 'cron'",
			},
			{
				Name:        "timezone",
				DisplayName: "Timezone",
				Type:        "options",
				Required:    false,
				Default:     "UTC",
				Description: "Timezone for schedule execution",
				Options: []wtypes.ParamOption{
					{Name: "UTC", Value: "UTC"},
					{Name: "US/Eastern", Value: "America/New_York"},
					{Name: "US/Pacific", Value: "America/Los_Angeles"},
					{Name: "US/Central", Value: "America/Chicago"},
					{Name: "Europe/London", Value: "Europe/London"},
					{Name: "Europe/Paris", Value: "Europe/Paris"},
					{Name: "Asia/Tokyo", Value: "Asia/Tokyo"},
					{Name: "Asia/Shanghai", Value: "Asia/Shanghai"},
					{Name: "Asia/Kolkata", Value: "Asia/Kolkata"},
					{Name: "Australia/Sydney", Value: "Australia/Sydney"},
				},
			},
			{
				Name:        "start_time",
				DisplayName: "Start Time",
				Type:        "string",
				Required:    false,
				Description: "When to start the schedule (ISO 8601 format)",
				Placeholder: "2024-01-01T00:00:00Z",
			},
			{
				Name:        "end_time",
				DisplayName: "End Time",
				Type:        "string",
				Required:    false,
				Description: "When to stop the schedule (ISO 8601 format)",
				Placeholder: "2024-12-31T23:59:59Z",
			},
		},
	}
}
