package models

import (
	"time"

	"github.com/google/uuid"
)

// AuditAction constants
const (
	AuditActionCreate     = "create"
	AuditActionUpdate     = "update"
	AuditActionDelete     = "delete"
	AuditActionExecute    = "execute"
	AuditActionActivate   = "activate"
	AuditActionDeactivate = "deactivate"
	AuditActionLogin      = "login"
	AuditActionLogout     = "logout"
	AuditActionInvite     = "invite"
	AuditActionExport     = "export"
	AuditActionImport     = "import"
)

// AuditLog tracks all user actions for compliance
type AuditLog struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID  uuid.UUID  `gorm:"type:uuid;index;not null" json:"workspace_id"`
	UserID       uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`
	Action       string     `gorm:"size:30;not null;index" json:"action"`
	ResourceType string     `gorm:"size:50;not null;index" json:"resource_type"` // workflow, credential, etc.
	ResourceID   *uuid.UUID `gorm:"type:uuid;index" json:"resource_id,omitempty"`
	ResourceName *string    `gorm:"size:255" json:"resource_name,omitempty"`
	OldValue     JSON       `gorm:"type:jsonb" json:"old_value,omitempty"` // Before state
	NewValue     JSON       `gorm:"type:jsonb" json:"new_value,omitempty"` // After state
	Metadata     JSON       `gorm:"type:jsonb" json:"metadata,omitempty"`  // Additional context
	IPAddress    *string    `gorm:"size:45" json:"ip_address,omitempty"`
	UserAgent    *string    `gorm:"size:500" json:"user_agent,omitempty"`
	CreatedAt    time.Time  `gorm:"index" json:"created_at"`

	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
	User      User      `gorm:"foreignKey:UserID" json:"-"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
