package observability

import (
	"time"

	"github.com/google/uuid"
)

// WorkflowMetrics represents aggregated metrics for a workflow
type WorkflowMetrics struct {
	WorkflowID   uuid.UUID `json:"workflow_id"`
	WorkflowName string    `json:"workflow_name"`
	Period       string    `json:"period"` // hour, day, week, month

	// Execution counts
	TotalExecutions      int64 `json:"total_executions"`
	SuccessfulExecutions int64 `json:"successful_executions"`
	FailedExecutions     int64 `json:"failed_executions"`
	CanceledExecutions   int64 `json:"canceled_executions"`
	RunningExecutions    int64 `json:"running_executions"`

	// Rates
	SuccessRate float64 `json:"success_rate"`
	ErrorRate   float64 `json:"error_rate"`

	// Duration stats (milliseconds)
	AvgDuration float64 `json:"avg_duration_ms"`
	MinDuration int64   `json:"min_duration_ms"`
	MaxDuration int64   `json:"max_duration_ms"`
	P50Duration float64 `json:"p50_duration_ms"`
	P95Duration float64 `json:"p95_duration_ms"`
	P99Duration float64 `json:"p99_duration_ms"`

	// Node stats
	AvgNodesExecuted float64 `json:"avg_nodes_executed"`
	MostFailedNode   string  `json:"most_failed_node,omitempty"`
	SlowestNode      string  `json:"slowest_node,omitempty"`

	// Cost
	TotalCreditsUsed float64 `json:"total_credits_used"`
	AvgCreditsPerRun float64 `json:"avg_credits_per_run"`
	EstimatedCostUSD float64 `json:"estimated_cost_usd"`

	// Time range
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// WorkspaceMetrics represents aggregated metrics for a workspace
type WorkspaceMetrics struct {
	WorkspaceID   uuid.UUID `json:"workspace_id"`
	WorkspaceName string    `json:"workspace_name"`
	Period        string    `json:"period"`

	// Workflow counts
	TotalWorkflows  int64 `json:"total_workflows"`
	ActiveWorkflows int64 `json:"active_workflows"`

	// Execution totals
	TotalExecutions      int64   `json:"total_executions"`
	SuccessfulExecutions int64   `json:"successful_executions"`
	FailedExecutions     int64   `json:"failed_executions"`
	SuccessRate          float64 `json:"success_rate"`

	// Duration
	AvgDuration float64 `json:"avg_duration_ms"`

	// Top performers
	TopWorkflowsByExecutions []WorkflowSummary `json:"top_workflows_by_executions"`
	TopWorkflowsByErrors     []WorkflowSummary `json:"top_workflows_by_errors"`
	TopWorkflowsByCost       []WorkflowSummary `json:"top_workflows_by_cost"`

	// Resource usage
	TotalCreditsUsed   float64 `json:"total_credits_used"`
	CreditsRemaining   float64 `json:"credits_remaining"`
	EstimatedCostUSD   float64 `json:"estimated_cost_usd"`
	APICallsCount      int64   `json:"api_calls_count"`
	DataProcessedBytes int64   `json:"data_processed_bytes"`

	// Trends
	ExecutionTrend []TrendPoint `json:"execution_trend"`
	ErrorTrend     []TrendPoint `json:"error_trend"`

	// Time range
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// WorkflowSummary represents a summary of a workflow for rankings
type WorkflowSummary struct {
	WorkflowID   uuid.UUID `json:"workflow_id"`
	WorkflowName string    `json:"workflow_name"`
	Value        float64   `json:"value"`
	Label        string    `json:"label,omitempty"`
}

// TrendPoint represents a point in a time series
type TrendPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Label     string    `json:"label,omitempty"`
}

// NodeMetrics represents metrics for a specific node type
type NodeMetrics struct {
	NodeID      string  `json:"node_id"`
	NodeType    string  `json:"node_type"`
	NodeName    string  `json:"node_name"`
	Executions  int64   `json:"executions"`
	Failures    int64   `json:"failures"`
	AvgDuration float64 `json:"avg_duration_ms"`
	ErrorRate   float64 `json:"error_rate"`
}

// ExecutionHeatmap represents execution distribution over time
type ExecutionHeatmap struct {
	Data []HeatmapCell `json:"data"`
}

// HeatmapCell represents a single cell in the heatmap
type HeatmapCell struct {
	DayOfWeek   int     `json:"day_of_week"` // 0=Sunday
	HourOfDay   int     `json:"hour_of_day"` // 0-23
	Count       int64   `json:"count"`
	AvgDuration float64 `json:"avg_duration_ms"`
}

// AlertRule represents an observability alert rule
type AlertRule struct {
	ID            uuid.UUID  `json:"id"`
	WorkspaceID   uuid.UUID  `json:"workspace_id"`
	WorkflowID    *uuid.UUID `json:"workflow_id,omitempty"`
	Name          string     `json:"name"`
	Description   string     `json:"description,omitempty"`
	Metric        string     `json:"metric"`   // error_rate, duration, failures
	Operator      string     `json:"operator"` // gt, lt, eq, gte, lte
	Threshold     float64    `json:"threshold"`
	Window        string     `json:"window"`   // 5m, 1h, 24h
	Channels      []string   `json:"channels"` // email, slack, webhook
	Enabled       bool       `json:"enabled"`
	LastTriggered *time.Time `json:"last_triggered,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// Alert represents a triggered alert
type Alert struct {
	ID          uuid.UUID  `json:"id"`
	RuleID      uuid.UUID  `json:"rule_id"`
	WorkspaceID uuid.UUID  `json:"workspace_id"`
	WorkflowID  *uuid.UUID `json:"workflow_id,omitempty"`
	Metric      string     `json:"metric"`
	Value       float64    `json:"value"`
	Threshold   float64    `json:"threshold"`
	Message     string     `json:"message"`
	Status      string     `json:"status"` // triggered, acknowledged, resolved
	TriggeredAt time.Time  `json:"triggered_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
}
