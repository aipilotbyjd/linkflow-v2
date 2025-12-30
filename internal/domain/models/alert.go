package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AlertType constants
const (
	AlertTypeEmail   = "email"
	AlertTypeSlack   = "slack"
	AlertTypeWebhook = "webhook"
	AlertTypeSMS     = "sms"
)

// AlertTrigger constants
const (
	AlertTriggerOnFailure    = "on_failure"
	AlertTriggerOnSuccess    = "on_success"
	AlertTriggerOnTimeout    = "on_timeout"
	AlertTriggerOnLongRun    = "on_long_run"
	AlertTriggerOnQuotaLimit = "on_quota_limit"
)

// Alert represents configurable alerts for workflows/executions
type Alert struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID  uuid.UUID      `gorm:"type:uuid;index;not null" json:"workspace_id"`
	WorkflowID   *uuid.UUID     `gorm:"type:uuid;index" json:"workflow_id,omitempty"` // nil = workspace-wide
	CreatedBy    uuid.UUID      `gorm:"type:uuid;not null" json:"created_by"`
	Name         string         `gorm:"size:100;not null" json:"name"`
	Type         string         `gorm:"size:20;not null" json:"type"`      // email, slack, webhook, sms
	Trigger      string         `gorm:"size:30;not null" json:"trigger"`   // on_failure, on_success, etc.
	Config       JSON           `gorm:"type:jsonb;not null" json:"config"` // Type-specific config
	Conditions   JSON           `gorm:"type:jsonb" json:"conditions"`      // Additional conditions
	CooldownMins int            `gorm:"default:5" json:"cooldown_mins"`    // Min minutes between alerts
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	LastFiredAt  *time.Time     `json:"last_fired_at,omitempty"`
	FireCount    int            `gorm:"default:0" json:"fire_count"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
	Workflow  *Workflow `gorm:"foreignKey:WorkflowID" json:"-"`
	Creator   User      `gorm:"foreignKey:CreatedBy" json:"-"`
}

func (Alert) TableName() string {
	return "alerts"
}

// AlertLog tracks fired alerts
type AlertLog struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AlertID     uuid.UUID  `gorm:"type:uuid;index;not null" json:"alert_id"`
	ExecutionID *uuid.UUID `gorm:"type:uuid;index" json:"execution_id,omitempty"`
	Status      string     `gorm:"size:20;not null" json:"status"` // sent, failed, skipped
	Message     string     `gorm:"type:text" json:"message"`
	Response    *string    `gorm:"type:text" json:"response,omitempty"`
	Error       *string    `gorm:"type:text" json:"error,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`

	Alert     Alert      `gorm:"foreignKey:AlertID" json:"-"`
	Execution *Execution `gorm:"foreignKey:ExecutionID" json:"-"`
}

func (AlertLog) TableName() string {
	return "alert_logs"
}
