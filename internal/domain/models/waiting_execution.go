package models

import (
	"time"

	"github.com/google/uuid"
)

// WaitingExecution represents an execution paused waiting for external resume
type WaitingExecution struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ExecutionID   uuid.UUID  `gorm:"type:uuid;index;not null" json:"execution_id"`
	WorkflowID    uuid.UUID  `gorm:"type:uuid;index;not null" json:"workflow_id"`
	WorkspaceID   uuid.UUID  `gorm:"type:uuid;index;not null" json:"workspace_id"`
	NodeID        string     `gorm:"size:100;not null" json:"node_id"`
	ResumeToken   string     `gorm:"size:255;uniqueIndex;not null" json:"resume_token"`
	ResumeType    string     `gorm:"size:50;not null" json:"resume_type"` // webhook, manual, timeout
	WebhookPath   *string    `gorm:"size:255;index" json:"webhook_path,omitempty"`
	TimeoutAt     *time.Time `json:"timeout_at,omitempty"`
	ResumedAt     *time.Time `json:"resumed_at,omitempty"`
	ResumeData    JSON       `gorm:"type:jsonb" json:"resume_data,omitempty"`
	ExecutionData JSON       `gorm:"type:jsonb" json:"execution_data,omitempty"` // Serialized execution state
	Status        string     `gorm:"size:20;not null;default:waiting" json:"status"` // waiting, resumed, expired
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`

	Execution Execution `gorm:"foreignKey:ExecutionID" json:"-"`
	Workflow  Workflow  `gorm:"foreignKey:WorkflowID" json:"-"`
}

func (WaitingExecution) TableName() string {
	return "waiting_executions"
}

// Template represents a workflow template
type Template struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Description *string   `gorm:"type:text" json:"description,omitempty"`
	Category    string    `gorm:"size:100;index" json:"category"`
	Tags        StringArray `gorm:"type:text[]" json:"tags"`
	IconURL     *string   `gorm:"size:500" json:"icon_url,omitempty"`
	Nodes       JSONArray `gorm:"type:jsonb;not null" json:"nodes"`
	Connections JSONArray `gorm:"type:jsonb;not null" json:"connections"`
	Settings    JSON      `gorm:"type:jsonb" json:"settings"`
	Variables   JSON      `gorm:"type:jsonb" json:"variables,omitempty"` // Required variables to fill
	IsPublic    bool      `gorm:"default:false" json:"is_public"`
	IsFeatured  bool      `gorm:"default:false" json:"is_featured"`
	UseCount    int       `gorm:"default:0" json:"use_count"`
	CreatedBy   *uuid.UUID `gorm:"type:uuid" json:"created_by,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Creator *User `gorm:"foreignKey:CreatedBy" json:"-"`
}

func (Template) TableName() string {
	return "templates"
}

// OAuthState stores OAuth flow state
type OAuthState struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	State       string    `gorm:"size:255;uniqueIndex;not null" json:"-"`
	UserID      uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	WorkspaceID uuid.UUID `gorm:"type:uuid;not null" json:"workspace_id"`
	Provider    string    `gorm:"size:50;not null" json:"provider"`
	RedirectURL string    `gorm:"size:500;not null" json:"redirect_url"`
	Scopes      StringArray `gorm:"type:text[]" json:"scopes"`
	ExpiresAt   time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`

	User      User      `gorm:"foreignKey:UserID" json:"-"`
	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
}

func (OAuthState) TableName() string {
	return "oauth_states"
}

// PinnedData stores pinned test data for nodes
type PinnedData struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkflowID  uuid.UUID `gorm:"type:uuid;index;not null" json:"workflow_id"`
	WorkspaceID uuid.UUID `gorm:"type:uuid;index;not null" json:"workspace_id"`
	NodeID      string    `gorm:"size:100;not null" json:"node_id"`
	Name        string    `gorm:"size:255" json:"name"`
	Data        JSON      `gorm:"type:jsonb;not null" json:"data"`
	CreatedBy   uuid.UUID `gorm:"type:uuid;not null" json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Workflow  Workflow `gorm:"foreignKey:WorkflowID" json:"-"`
	Creator   User     `gorm:"foreignKey:CreatedBy" json:"-"`
}

func (PinnedData) TableName() string {
	return "pinned_data"
}

// ErrorWorkflow constants
const (
	ErrorTriggerOnFailure = "on_failure"
	ErrorTriggerOnTimeout = "on_timeout"
	ErrorTriggerOnAll     = "on_all"
)

// BinaryData represents stored binary data for executions
type BinaryData struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ExecutionID uuid.UUID `gorm:"type:uuid;index;not null" json:"execution_id"`
	WorkspaceID uuid.UUID `gorm:"type:uuid;index;not null" json:"workspace_id"`
	NodeID      string    `gorm:"size:100" json:"node_id,omitempty"`
	FileName    string    `gorm:"size:255;not null" json:"file_name"`
	MimeType    string    `gorm:"size:100;not null" json:"mime_type"`
	Size        int64     `gorm:"not null" json:"size"`
	StoragePath string    `gorm:"size:500;not null" json:"-"` // Internal storage path
	StorageType string    `gorm:"size:50;not null" json:"storage_type"` // local, s3, gcs
	Checksum    string    `gorm:"size:64" json:"checksum,omitempty"`
	Metadata    JSON      `gorm:"type:jsonb" json:"metadata,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`

	Execution Execution `gorm:"foreignKey:ExecutionID" json:"-"`
}

func (BinaryData) TableName() string {
	return "binary_data"
}
