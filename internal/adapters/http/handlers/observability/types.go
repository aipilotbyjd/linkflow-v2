package observability

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/observability"
)

// MetricsRequest represents a request for metrics
type MetricsRequest struct {
	Period    string `json:"period"` // hour, day, week, month
	StartTime string `json:"start_time,omitempty"`
	EndTime   string `json:"end_time,omitempty"`
}

// WorkflowMetricsResponse represents workflow metrics response
type WorkflowMetricsResponse struct {
	Metrics     observability.WorkflowMetrics `json:"metrics"`
	NodeMetrics []observability.NodeMetrics   `json:"node_metrics,omitempty"`
	Trend       []observability.TrendPoint    `json:"trend,omitempty"`
}

// WorkspaceMetricsResponse represents workspace metrics response
type WorkspaceMetricsResponse struct {
	Metrics observability.WorkspaceMetrics  `json:"metrics"`
	Heatmap *observability.ExecutionHeatmap `json:"heatmap,omitempty"`
}

// AlertRuleRequest represents an alert rule create/update request
type AlertRuleRequest struct {
	Name        string   `json:"name" validate:"required"`
	Description string   `json:"description,omitempty"`
	WorkflowID  *string  `json:"workflow_id,omitempty"`
	Metric      string   `json:"metric" validate:"required"`
	Operator    string   `json:"operator" validate:"required"`
	Threshold   float64  `json:"threshold" validate:"required"`
	Window      string   `json:"window" validate:"required"`
	Channels    []string `json:"channels" validate:"required"`
	Enabled     bool     `json:"enabled"`
}

// AlertRuleResponse represents an alert rule response
type AlertRuleResponse struct {
	ID          string   `json:"id"`
	WorkspaceID string   `json:"workspace_id"`
	WorkflowID  *string  `json:"workflow_id,omitempty"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Metric      string   `json:"metric"`
	Operator    string   `json:"operator"`
	Threshold   float64  `json:"threshold"`
	Window      string   `json:"window"`
	Channels    []string `json:"channels"`
	Enabled     bool     `json:"enabled"`
	CreatedAt   string   `json:"created_at"`
}

// AlertResponse represents an alert response
type AlertResponse struct {
	ID          string  `json:"id"`
	RuleID      string  `json:"rule_id"`
	WorkflowID  *string `json:"workflow_id,omitempty"`
	Metric      string  `json:"metric"`
	Value       float64 `json:"value"`
	Threshold   float64 `json:"threshold"`
	Message     string  `json:"message"`
	Status      string  `json:"status"`
	TriggeredAt string  `json:"triggered_at"`
}

// Service interface
type Service interface {
	GetWorkflowMetrics(ctx context.Context, workflowID uuid.UUID, period string) (*observability.WorkflowMetrics, error)
	GetWorkspaceMetrics(ctx context.Context, workspaceID uuid.UUID, period string) (*observability.WorkspaceMetrics, error)
	GetExecutionHeatmap(ctx context.Context, workspaceID uuid.UUID, period string) (*observability.ExecutionHeatmap, error)
	CreateAlertRule(ctx context.Context, workspaceID uuid.UUID, req AlertRuleRequest) (*observability.AlertRule, error)
	ListAlertRules(ctx context.Context, workspaceID uuid.UUID) ([]observability.AlertRule, error)
	ListAlerts(ctx context.Context, workspaceID uuid.UUID, status string) ([]observability.Alert, error)
}
