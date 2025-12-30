package models

import (
	"time"

	"github.com/google/uuid"
)

// WorkspaceAnalytics stores aggregated analytics
type WorkspaceAnalytics struct {
	ID                   uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID          uuid.UUID `gorm:"type:uuid;index;not null" json:"workspace_id"`
	Date                 time.Time `gorm:"type:date;index;not null" json:"date"`
	ExecutionsTotal      int       `gorm:"default:0" json:"executions_total"`
	ExecutionsSuccess    int       `gorm:"default:0" json:"executions_success"`
	ExecutionsFailed     int       `gorm:"default:0" json:"executions_failed"`
	ExecutionsCancelled  int       `gorm:"default:0" json:"executions_cancelled"`
	AvgExecutionDuration int       `gorm:"default:0" json:"avg_execution_duration_ms"`
	TotalOperations      int       `gorm:"default:0" json:"total_operations"`
	CreditsUsed          int       `gorm:"default:0" json:"credits_used"`
	WebhooksReceived     int       `gorm:"default:0" json:"webhooks_received"`
	SchedulesTriggered   int       `gorm:"default:0" json:"schedules_triggered"`
	ErrorsByType         JSON      `gorm:"type:jsonb" json:"errors_by_type"` // {error_type: count}
	TopWorkflows         JSON      `gorm:"type:jsonb" json:"top_workflows"`  // [{workflow_id, count}]
	TopErrors            JSON      `gorm:"type:jsonb" json:"top_errors"`     // [{error, count}]
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`

	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
}

func (WorkspaceAnalytics) TableName() string {
	return "workspace_analytics"
}

// WorkflowAnalytics stores per-workflow analytics
type WorkflowAnalytics struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkflowID        uuid.UUID `gorm:"type:uuid;index;not null" json:"workflow_id"`
	WorkspaceID       uuid.UUID `gorm:"type:uuid;index;not null" json:"workspace_id"`
	Date              time.Time `gorm:"type:date;index;not null" json:"date"`
	ExecutionsTotal   int       `gorm:"default:0" json:"executions_total"`
	ExecutionsSuccess int       `gorm:"default:0" json:"executions_success"`
	ExecutionsFailed  int       `gorm:"default:0" json:"executions_failed"`
	AvgDurationMs     int       `gorm:"default:0" json:"avg_duration_ms"`
	MinDurationMs     int       `gorm:"default:0" json:"min_duration_ms"`
	MaxDurationMs     int       `gorm:"default:0" json:"max_duration_ms"`
	P95DurationMs     int       `gorm:"default:0" json:"p95_duration_ms"`
	CreditsUsed       int       `gorm:"default:0" json:"credits_used"`
	SuccessRate       float64   `gorm:"default:0" json:"success_rate"`
	NodePerformance   JSON      `gorm:"type:jsonb" json:"node_performance"` // {node_id: {avg_ms, success_rate}}
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`

	Workflow  Workflow  `gorm:"foreignKey:WorkflowID" json:"-"`
	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
}

func (WorkflowAnalytics) TableName() string {
	return "workflow_analytics"
}
