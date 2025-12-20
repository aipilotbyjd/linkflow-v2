package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WebhookEndpoint struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkflowID  uuid.UUID      `gorm:"type:uuid;index;not null" json:"workflow_id"`
	WorkspaceID uuid.UUID      `gorm:"type:uuid;index;not null" json:"workspace_id"`
	NodeID      string         `gorm:"size:100;not null" json:"node_id"`
	Path        string         `gorm:"size:255;uniqueIndex;not null" json:"path"`
	Method      string         `gorm:"size:10;default:POST" json:"method"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	Secret      *string        `gorm:"size:255" json:"-"`
	LastCalledAt *time.Time    `json:"last_called_at,omitempty"`
	CallCount   int            `gorm:"default:0" json:"call_count"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Workflow  Workflow  `gorm:"foreignKey:WorkflowID" json:"-"`
	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
}

func (WebhookEndpoint) TableName() string {
	return "webhook_endpoints"
}

type WebhookLog struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	EndpointID uuid.UUID `gorm:"type:uuid;index;not null" json:"endpoint_id"`
	Method     string    `gorm:"size:10;not null" json:"method"`
	Path       string    `gorm:"size:255;not null" json:"path"`
	Headers    JSON      `gorm:"type:jsonb" json:"headers,omitempty"`
	Body       *string   `gorm:"type:text" json:"body,omitempty"`
	IPAddress  *string   `gorm:"size:45" json:"ip_address,omitempty"`
	StatusCode int       `gorm:"not null" json:"status_code"`
	Response   *string   `gorm:"type:text" json:"response,omitempty"`
	DurationMs int       `gorm:"not null" json:"duration_ms"`
	CreatedAt  time.Time `json:"created_at"`

	Endpoint WebhookEndpoint `gorm:"foreignKey:EndpointID" json:"-"`
}

func (WebhookLog) TableName() string {
	return "webhook_logs"
}
